package algorithm

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

var fogNodes []*FogNode
var edgeNodes []*EdgeNode

//This function builds a topology with the given parameters and then measures memory use via bash.
//Very similar to the GenerateNodes function, but they're immediately added to the topology.
func GetMemoryForClusters(numEdge int, numFog int, pingDiff float64, slaMaxPing float32, checkResources bool) StatsLine {
	clusterer = (&LiveCluster{}).Init()

	//Generates a number of fog nodes in random positions between 0..1200 and 0..750, excluding a 100 unit border on all sides.
	//The resource numbers are pretty random, they're not very important for my tests, they just need to exist.
	for i := 0; i < numFog; i++ {
		nodeResources := make(map[Resource]int)
		nodeResources[CPUShares] = 2000
		nodeResources[Memory] = 8000
		nodeResources[Network] = 10000000

		node := FogNode{
			X:              float64(100 + rand.Int31n(1000)),
			Y:              float64(100 + rand.Int31n(550)),
			ResourcesTotal: nodeResources,
		}
		node.Active = false
		node.ResourcesUsed = make(map[Resource]int)
		clusterer.FogNodes = append(clusterer.FogNodes, &node)
	}

	//maxRelDiff determines how much ping can vary with distance.
	//example: distance is 100, pingDiff is 20. maxRelDiff will be 0.2 so ping can be anywhere from 80 to 120.
	//attempt to simulate dependency of ping on other things than pure distance.
	maxRelDiff := pingDiff / 100
	//number of generated "cores" of edge nodes varies with total number of edge nodes
	numGenClusters := int(2 + 2*math.Log10(float64(numEdge)))
	//keeps track of how many edge nodes are left to generate
	edgeNodesLeft := numEdge

	//Generates numGenCluster "cores" of edge nodes, centered on clusterX, clusterY (see a few lines below), with a
	//random max radius. All edge nodes for a core are generated with random r < max radius and theta 0...2pi.
	//This method makes the middle of each core slightly denser than the outer regions, simulating "cities" or "towns". They can overlap!
	for curCluster := 0; curCluster < numGenClusters; curCluster++ {
		nodesInCluster := int(rand.Int31n(int32(2*numEdge/numGenClusters)) + 1)
		nodesInCluster = int(math.Min(float64(nodesInCluster), float64(edgeNodesLeft)))

		if curCluster == (numGenClusters - 1) {
			if edgeNodesLeft <= 0 {
				nodesInCluster = 1
			} else {
				nodesInCluster = edgeNodesLeft
			}
		}

		edgeNodesLeft -= nodesInCluster

		clusterSize := 50 + int(rand.Int31n(150))
		clusterX := clusterSize + int(rand.Int31n(int32(1200-2*clusterSize)))
		clusterY := clusterSize + int(rand.Int31n(int32(750-2*clusterSize)))

		for node := 0; node < nodesInCluster; node++ {
			r := float64(rand.Int31n(int32(clusterSize)))
			theta := rand.Float64() * 2 * math.Pi

			nodeX := clusterX + int(math.Cos(theta)*r)
			nodeY := clusterY + int(math.Sin(theta)*r)

			fogNodePings := make(map[*FogNode]float32)

			//Generates a ping from the current edge node to each fog node.
			//Horrible on memory, remove if you don't need this type of metric.
			for _, fn := range clusterer.FogNodes {
				dist := math.Sqrt(math.Pow(fn.X-float64(nodeX), 2) + math.Pow(fn.Y-float64(nodeY), 2))
				ping := ((1 - maxRelDiff) + rand.Float64()*(2*maxRelDiff)) * dist
				fogNodePings[fn] = float32(ping)
			}

			nodes, pings := SortNodePings(fogNodePings)
			node := EdgeNode{
				X:           float64(nodeX),
				Y:           float64(nodeY),
				Pings:       pings,
				SortedPings: nodes,
				Name:        fmt.Sprintf("%d-%d", curCluster, node),
			}
			clusterer.ProcessNode(slaMaxPing, &node, checkResources)
		}
	}
	pid := os.Getpid()
	memLine, _ := ExecCmdBash(fmt.Sprintf("cat /proc/%d/statm", pid))
	//fmt.Printf("%s\n", memLine)
	memParts := strings.Split(memLine, " ")
	mem, _ := strconv.Atoi(memParts[1])
	//fmt.Printf("%d\n", mem)
	stats := GetStats(numEdge, numFog)
	stats.Metrics["Memory"] = float64(mem * 4)
	stats.MNames = []string{"Memory"}
	return stats
}

func ExecCmdBash(dfCmd string) (string, error) {
	cmd := exec.Command("sh", "-c", dfCmd)
	stdout, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return "", err
	}
	return string(stdout), nil
}

//This function generates a number of fog nodes at random positions, and a number of "cores" of edge nodes representing "towns".
//More or less the same as the GetMemoryForClusters function above, but this one keeps the generated topology around for use in the ClusterIncremental...
//functions and DeleteNodes function below.
func GenerateNodes(numEdge int, numFog int, pingDiff float64) {
	fogNodes = []*FogNode{}
	edgeNodes = []*EdgeNode{}

	for i := 0; i < numFog; i++ {
		nodeResources := make(map[Resource]int)
		nodeResources[CPUShares] = 2000
		nodeResources[Memory] = 8000
		nodeResources[Network] = 10000000

		node := FogNode{
			X:              float64(100 + rand.Int31n(1000)),
			Y:              float64(100 + rand.Int31n(550)),
			ResourcesTotal: nodeResources,
		}
		fogNodes = append(fogNodes, &node)
	}

	maxRelDiff := pingDiff / 100
	numGenClusters := int(2 + 2*math.Log10(float64(numEdge)))
	edgeNodesLeft := numEdge

	for curCluster := 0; curCluster < numGenClusters; curCluster++ {
		nodesInCluster := int(rand.Int31n(int32(2*numEdge/numGenClusters)) + 1)
		nodesInCluster = int(math.Min(float64(nodesInCluster), float64(edgeNodesLeft)))

		if curCluster == (numGenClusters - 1) {
			if edgeNodesLeft <= 0 {
				nodesInCluster = 1
			} else {
				nodesInCluster = edgeNodesLeft
			}
		}

		edgeNodesLeft -= nodesInCluster

		clusterSize := 50 + int(rand.Int31n(150))
		clusterX := clusterSize + int(rand.Int31n(int32(1200-2*clusterSize)))
		clusterY := clusterSize + int(rand.Int31n(int32(750-2*clusterSize)))

		for node := 0; node < nodesInCluster; node++ {
			r := float64(rand.Int31n(int32(clusterSize)))
			theta := rand.Float64() * 2 * math.Pi

			nodeX := clusterX + int(math.Cos(theta)*r)
			nodeY := clusterY + int(math.Sin(theta)*r)

			fogNodePings := make(map[*FogNode]float32)

			for _, fn := range fogNodes {
				dist := math.Sqrt(math.Pow(fn.X-float64(nodeX), 2) + math.Pow(fn.Y-float64(nodeY), 2))
				ping := ((1 - maxRelDiff) + rand.Float64()*(2*maxRelDiff)) * dist
				fogNodePings[fn] = float32(ping)
			}

			nodes, pings := SortNodePings(fogNodePings)
			node := EdgeNode{
				X:           float64(nodeX),
				Y:           float64(nodeY),
				Pings:       pings,
				SortedPings: nodes,
				Name:        fmt.Sprintf("%d-%d", curCluster, node),
			}
			edgeNodes = append(edgeNodes, &node)
		}
	}
}

var clusterer *LiveCluster

//Deletes a number of nodes from the LiveCluster and returns timing information.
func DeleteNodes(slaMaxPing float32, numNodes int, checkResources bool) StatsLine {
	start := time.Now()
	numEdge := len(edgeNodes)
	for i := 0; i < numNodes; i++ {
		clusterer.RemoveNode(slaMaxPing, edgeNodes[i], checkResources)
	}
	stop := time.Now()
	timeTakenMs := float64(stop.Sub(start).Nanoseconds()) / 1000000
	stats := GetStats(numEdge, len(fogNodes))
	stats.Metrics["TimeTakenMs"] = timeTakenMs
	stats.MNames = []string{"TimeTakenMs"}
	return stats
}

//Performs an incremental cluster of all edge nodes and returns timing information.
func ClusterIncrementalTimeStats(slaMaxPing float32, checkResources bool) StatsLine {
	start := time.Now()

	ClusterIncremental(slaMaxPing, checkResources)

	stop := time.Now()
	timeTakenMs := float64(stop.Sub(start).Nanoseconds()) / 1000000
	stats := GetStats(len(edgeNodes), len(fogNodes))
	stats.Metrics["TimeTakenMs"] = timeTakenMs
	stats.MNames = []string{"TimeTakenMs"}
	return stats
}

//Performs an incremental cluster of all edge nodes and returns cluster stats (avg ping, theoretical min ping, ...)
func ClusterIncrementalClusterStats(slaMaxPing float32, checkResources bool) StatsLine {
	ClusterIncremental(slaMaxPing, checkResources)

	numClusters := len(clusterer.ActiveClusters)
	var avgDist, avgMinDist float32

	for en, fn := range clusterer.ClusteredNodes {
		avgDist += en.GetPing(fn)
		avgMinDist += en.GetPing(en.SortedPings[0])
	}
	avgDist /= float32(len(clusterer.ClusteredNodes))
	avgMinDist /= float32(len(clusterer.ClusteredNodes))

	stats := GetStats(len(edgeNodes), len(fogNodes))
	stats.Stats["NumClusters"] = float64(numClusters)
	stats.SNames = []string{"NumClusters"}
	stats.Metrics["AvgDist"] = float64(avgDist)
	stats.Metrics["AvgMinDist"] = float64(avgMinDist)
	stats.MNames = []string{"AvgDist", "AvgMinDist"}
	return stats
}

//Performs an incremental cluster with dynamic selection of fog nodes disabled.
//Instead, a number of random fog nodes are preselected to cluster all edge nodes to.
//Returns cluster stats.
func ClusterIncrementalStaticFognodes(slaMaxPing float32, checkResources bool, clusters int) StatsLine {
	if clusters == 0 {
		clusters = len(clusterer.ActiveClusters)
	}
	clusterer.InitStaticClusters(clusters)
	for _, en := range edgeNodes {
		clusterer.ProcessNode(slaMaxPing, en, checkResources)
	}

	numClusters := len(clusterer.ActiveClusters)
	var avgDist, avgMinDist float32

	for en, fn := range clusterer.ClusteredNodes {
		avgDist += en.GetPing(fn)
		avgMinDist += en.GetPing(en.SortedPings[0])
	}
	avgDist /= float32(len(clusterer.ClusteredNodes))
	avgMinDist /= float32(len(clusterer.ClusteredNodes))

	stats := GetStats(len(edgeNodes), len(fogNodes))
	stats.Stats["NumClusters"] = float64(numClusters)
	stats.SNames = []string{"NumClusters"}
	stats.Metrics["AvgDist"] = float64(avgDist)
	stats.Metrics["AvgMinDist"] = float64(avgMinDist)
	stats.MNames = []string{"AvgDist", "AvgMinDist"}
	return stats
}

func ClosestClusterIncrementalStaticFognodes(slaMaxPing float32, checkResources bool, clusters int) StatsLine {
	if clusters == 0 {
		clusters = len(clusterer.ActiveClusters)
	}
	clusterer.InitStaticClusters(clusters)
	//for _, en := range edgeNodes {
	clusterer.AssignClosest(edgeNodes, checkResources)
	//}

	numClusters := len(clusterer.ActiveClusters)
	var avgDist, avgMinDist float32

	for en, fn := range clusterer.ClusteredNodes {
		avgDist += en.GetPing(fn)
		avgMinDist += en.GetPing(en.SortedPings[0])
	}
	avgDist /= float32(len(clusterer.ClusteredNodes))
	avgMinDist /= float32(len(clusterer.ClusteredNodes))

	stats := GetStats(len(edgeNodes), len(fogNodes))
	stats.Stats["NumClusters"] = float64(numClusters)
	stats.SNames = []string{"NumClusters"}
	stats.Metrics["AvgDist"] = float64(avgDist)
	stats.Metrics["AvgMinDist"] = float64(avgMinDist)
	stats.MNames = []string{"AvgDist", "AvgMinDist"}
	return stats
}

//This function simply adds all pre-generated edge nodes to the LiveCluster to build a service topology.
func ClusterIncremental(slaMaxPing float32, checkResources bool) {
	clusterer = (&LiveCluster{}).Init()
	for _, fn := range fogNodes {
		fn.Active = false
		fn.ResourcesUsed = make(map[Resource]int)
	}
	clusterer.FogNodes = fogNodes

	for _, en := range edgeNodes {
		clusterer.ProcessNode(slaMaxPing, en, checkResources)
	}
}

func GetStats(eNodes int, fNodes int) StatsLine {
	line := StatsLine{}.Init()

	line.EdgeNodes = eNodes
	line.FogNodes = fNodes
	return line
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//Again, a horrible golang mess. Because golang provides no real way to sort maps according to their key values,
//this thing had to be written.
func SortNodePings(pings map[*FogNode]float32) ([]*FogNode, []float32) {
	nodePings := []NodePing{}
	for fn, ping := range pings {
		nPing := NodePing{Node: fn, Ping: ping}
		nodePings = append(nodePings, nPing)
	}

	ping := func(p1, p2 *NodePing) bool {
		return p1.Ping < p2.Ping
	}
	By(ping).Sort(nodePings)

	nodes := []*FogNode{}
	npings := []float32{}
	for i := 0; i < len(nodePings); i++ {
		nodes = append(nodes, nodePings[i].Node)
		npings = append(npings, nodePings[i].Ping)
	}
	return nodes, npings
}

type NodePing struct {
	Node *FogNode
	Ping float32
}

type By func(p1, p2 *NodePing) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(pings []NodePing) {
	ps := &pingSorter{
		pings: pings,
		by:    by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

type pingSorter struct {
	pings []NodePing
	by    func(p1, p2 *NodePing) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *pingSorter) Len() int {
	return len(s.pings)
}

// Swap is part of sort.Interface.
func (s *pingSorter) Swap(i, j int) {
	s.pings[i], s.pings[j] = s.pings[j], s.pings[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *pingSorter) Less(i, j int) bool {
	return s.by(&s.pings[i], &s.pings[j])
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func SortENPings(pings map[float32]*EdgeNode) ([]*EdgeNode, []float32) {
	nodePings := []ENodePing{}
	for ping, fn := range pings {
		nPing := ENodePing{Node: fn, Ping: ping}
		nodePings = append(nodePings, nPing)
	}

	ping := func(p1, p2 *ENodePing) bool {
		return p1.Ping < p2.Ping
	}
	ENBy(ping).Sort(nodePings)

	nodes := []*EdgeNode{}
	npings := []float32{}
	for i := 0; i < len(nodePings); i++ {
		nodes = append(nodes, nodePings[i].Node)
		npings = append(npings, nodePings[i].Ping)
	}
	return nodes, npings
}

type ENodePing struct {
	Node *EdgeNode
	Ping float32
}

type ENBy func(p1, p2 *ENodePing) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by ENBy) Sort(pings []ENodePing) {
	ps := &ePingSorter{
		pings: pings,
		by:    by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

type ePingSorter struct {
	pings []ENodePing
	by    func(p1, p2 *ENodePing) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *ePingSorter) Len() int {
	return len(s.pings)
}

// Swap is part of sort.Interface.
func (s *ePingSorter) Swap(i, j int) {
	s.pings[i], s.pings[j] = s.pings[j], s.pings[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *ePingSorter) Less(i, j int) bool {
	return s.by(&s.pings[i], &s.pings[j])
}
