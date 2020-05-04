package algorithm

import (
	//"math/rand"
	"fmt"
	//"strings"
)

var Clusterer *LiveCluster
var ClusterInitialized func(clusterID string)
var ClusterUninitialized func(clusterID string)

func Init() {
	Clusterer = (&LiveCluster{}).Init()
}

//NodeNames may look strange here, but it's required to make some things O(1) and doesn't use all that much memory.
type LiveCluster struct {
	EdgeNodes              map[string]*EdgeNode
	ClusteredNodes         map[*EdgeNode]*FogNode
	ActiveClusters         map[*FogNode]map[string]*EdgeNode //[]*EdgeNode
	FogNodes               map[string]*FogNode
	ClusterCreationAllowed bool
}

func (lc *LiveCluster) Init() *LiveCluster {
	lc.ClusterCreationAllowed = true
	lc.EdgeNodes = make(map[string]*EdgeNode)
	lc.ClusteredNodes = make(map[*EdgeNode]*FogNode)
	lc.ActiveClusters = make(map[*FogNode]map[string]*EdgeNode) //[]*EdgeNode)
	lc.FogNodes = make(map[string]*FogNode)
	return lc
}

func (lc *LiveCluster) DumpStructure() {
	for _, fn := range lc.FogNodes {
		fmt.Printf("%s, Active %t\n", fn.ID, fn.Active)
	}
	for _, en := range lc.EdgeNodes {
		pings := ""
		for idx := 0; idx < len(en.SortedPings); idx++ {
			pings = fmt.Sprintf("%s;%s %f", pings, en.SortedPings[idx], en.Pings[idx])
		}
		fmt.Printf("%s, #pings %s\n", en.ID, pings)
	}
	for en, fn := range lc.ClusteredNodes {
		fmt.Printf("%s clustered to %s\n", en.ID, fn.ID)
	}
}

//if it already exists, update resources,
//if not add it
//no other actions required here
func (lc *LiveCluster) ProcessFogNode(id string, resources map[Resource]int, totalResources map[Resource]int) bool {
	node, contains := lc.FogNodes[id]
	if contains {
		node.ResourcesTotal = totalResources
		node.ResourcesUsed = resources
		return false
	} else {
		lc.FogNodes[id] = &FogNode{
			ID:             id,
			Active:         false,
			ResourcesUsed:  resources,
			ResourcesTotal: totalResources,
		}
		lc.DumpStructure()
		return true
	}
}

//if the edge node does not yet exist in the swirl, add it
//if it does, update its pings
//if it's also part of the service topology, update that first
func (lc *LiveCluster) ProcessEdgeNode(slaMaxPing float32, edgeNode *EdgeNode, checkResources bool) {
	_, clustered := lc.ClusteredNodes[edgeNode]
	//see if service topology needs updating
	if clustered {
		lc.updateDeployment(slaMaxPing, edgeNode, checkResources)
	}
	//see if it needs adding or just updating pings
	existing, containsNode := lc.EdgeNodes[edgeNode.ID]
	if containsNode {
		existing.Pings = edgeNode.Pings
		existing.SortedPings = edgeNode.SortedPings
	} else {
		lc.EdgeNodes[edgeNode.ID] = edgeNode
	}
	lc.DumpStructure()
}

//remove from the swirl
func (lc *LiveCluster) RemoveEdgeNode(slaMaxPing float32, nodeID string, checkResources bool) {
	edgeNode := lc.EdgeNodes[nodeID]
	lc.RemoveDeployment(slaMaxPing, edgeNode.ID, checkResources)
	delete(lc.EdgeNodes, nodeID)
	lc.DumpStructure()
}

//this basically only needs to add a node to the service topology
//it really really really should already be in the swirl
func (lc *LiveCluster) ProcessDeployment(slaMaxPing float32, nodeID string, checkResources bool) {
	fmt.Printf("Process deployment for %s\n", nodeID)
	lc.DumpStructure()
	edgeNode, _ := lc.EdgeNodes[nodeID]
	//if containsNode {
	//	lc.UpdateNode(slaMaxPing, edgeNode, checkResources)
	//} else {
	lc.addDeployment(slaMaxPing, edgeNode, checkResources)
	//}
	lc.DumpStructure()
}

func (lc *LiveCluster) updateDeployment(slaMaxPing float32, newNode *EdgeNode, checkResources bool) {
	fogNode, _ := lc.ClusteredNodes[newNode]
	existing, _ := lc.EdgeNodes[newNode.ID]
	oldPing := existing.GetPing(fogNode) //.Pings[fogNode]
	newPing := newNode.GetPing(fogNode)  //.Pings[fogNode]

	existing.Pings = newNode.Pings
	existing.SortedPings = newNode.SortedPings

	//have to do something about it if ping went over max SLA
	if oldPing < slaMaxPing && newPing > slaMaxPing {
		lc.RemoveDeployment(slaMaxPing, existing.ID, checkResources)
		lc.addDeployment(slaMaxPing, existing, checkResources)
	}
	lc.DumpStructure()
}

//force remove a fog node
func (lc *LiveCluster) RemoveFogNode(nodeID string, slaMaxPing float32) {
	fogNode := lc.FogNodes[nodeID]

	sameCluster := lc.ActiveClusters[fogNode]
	toAdd := make(map[string]*EdgeNode)
	//remove the remaining edge nodes
	for len(sameCluster) > 0 {
		_, node := getFirstItem(sameCluster)
		toAdd[node.ID] = node

		lc.RemoveFromCluster(node, fogNode)
	}

	//fogNode.Active = false
	//delete(lc.ActiveClusters, fogNode)
	//fogNode.ResourcesUsed = make(map[Resource]int)
	lc.UninitCluster(fogNode)
	delete(lc.FogNodes, nodeID)

	//add remaining edge nodes to a new fog node
	for id, _ := range toAdd {
		lc.ProcessDeployment(slaMaxPing, id, true)
	}

	lc.DumpStructure()
}

//only remove from the topology
func (lc *LiveCluster) RemoveDeployment(slaMaxPing float32, nodeID string, checkResources bool) {
	edgeNode := lc.EdgeNodes[nodeID]
	fogNode := lc.ClusteredNodes[edgeNode]
	lc.RemoveFromCluster(edgeNode, fogNode)

	sameCluster := lc.ActiveClusters[fogNode]
	if (checkResources && fogNode.IsUnderutilized()) || (!checkResources && len(sameCluster) <= 200) {
		//try to move the remaining edge nodes
		failed := false
		newClusters := make(map[*EdgeNode]*FogNode)

		for len(sameCluster) > 0 && !failed {
			_, node := getFirstItem(sameCluster)
			closestOther, otherPing := lc.GetClosestClusterExcept(node, fogNode, checkResources)
			if closestOther != nil {
				fPing := node.GetPing(fogNode)
				if fPing > slaMaxPing || (fPing < slaMaxPing && otherPing < slaMaxPing) {
					lc.RemoveFromCluster(node, fogNode)
					lc.addToCluster(node, closestOther)
					newClusters[node] = closestOther
				} else {
					failed = true
				}
			} else {
				failed = true
			}
		}

		if failed {
			for node, cluster := range newClusters {
				lc.RemoveFromCluster(node, cluster)
				lc.addToCluster(node, fogNode)
			}
		} else {
			//fogNode.Active = false
			//delete(lc.ActiveClusters, fogNode)
			//fogNode.ResourcesUsed = make(map[Resource]int)
			lc.UninitCluster(fogNode)
		}
	}
	lc.DumpStructure()
}

func getFirstItem(nodes map[string]*EdgeNode) (string, *EdgeNode) {
	for a, b := range nodes {
		return a, b
	}
	return "", nil
}

func (lc *LiveCluster) addDeployment(slaMaxPing float32, edgeNode *EdgeNode, checkResources bool) {
	if len(lc.ActiveClusters) == 0 {
		lc.createNewClusterFor(edgeNode, nil, checkResources)
	} else {
		closestCluster, ping := lc.GetClosestCluster(edgeNode, checkResources)
		if closestCluster == nil && lc.ClusterCreationAllowed {
			lc.createNewClusterFor(edgeNode, nil, checkResources)
		} else if ping > slaMaxPing && lc.ClusterCreationAllowed {
			lc.TryCreateNewClusterFor(edgeNode, closestCluster, checkResources)
		} else {
			lc.addToCluster(edgeNode, closestCluster)
		}
	}
}

func (lc *LiveCluster) TryCreateNewClusterFor(edgeNode *EdgeNode, closestCluster *FogNode, checkResources bool) {
	closestFogNode, _ := lc.GetClosestFogNode(edgeNode, checkResources)
	if closestFogNode == closestCluster {
		lc.addToCluster(edgeNode, closestCluster)
	} else {
		lc.createNewClusterFor(edgeNode, closestFogNode, checkResources)
	}
}

func (lc *LiveCluster) createNewClusterFor(edgeNode *EdgeNode, fogNode *FogNode, checkResources bool) {
	if fogNode == nil {
		fogNode, _ = lc.GetClosestFogNode(edgeNode, checkResources)
	}

	_, exists := lc.ActiveClusters[fogNode]
	if exists {
		lc.addToCluster(edgeNode, fogNode)
	} else {
		lc.InitCluster(fogNode)
		lc.addToCluster(edgeNode, fogNode)
	}
}

func (lc *LiveCluster) addToCluster(edgeNode *EdgeNode, cluster *FogNode) {
	lc.ClusteredNodes[edgeNode] = cluster
	lc.ActiveClusters[cluster][edgeNode.ID] = edgeNode
	//lc.EdgeNodes[edgeNode.ID] = edgeNode

	cluster.ResourcesUsed[CPUShares] += 2
	cluster.ResourcesUsed[Memory] += 2
	cluster.ResourcesUsed[Network] += 1000
}

func (lc *LiveCluster) RemoveFromCluster(edgeNode *EdgeNode, cluster *FogNode) {
	delete(lc.ClusteredNodes, edgeNode)
	delete(lc.ActiveClusters[cluster], edgeNode.ID)
	//delete(lc.EdgeNodes, edgeNode.ID)

	cluster.ResourcesUsed[CPUShares] -= 2
	cluster.ResourcesUsed[Memory] -= 2
	cluster.ResourcesUsed[Network] -= 1000
}

func (lc *LiveCluster) UninitCluster(cluster *FogNode) {
	cluster.Active = false
	delete(lc.ActiveClusters, cluster)

	//cluster.ResourcesUsed[CPUShares] = 200
	//cluster.ResourcesUsed[Memory] = 200
	//cluster.ResourcesUsed[Network] = 0
	ClusterUninitialized(cluster.ID)
}

func (lc *LiveCluster) InitCluster(cluster *FogNode) {
	cluster.Active = true
	lc.ActiveClusters[cluster] = make(map[string]*EdgeNode)

	//cluster.ResourcesUsed[CPUShares] = 200
	//cluster.ResourcesUsed[Memory] = 200
	//cluster.ResourcesUsed[Network] = 0
	ClusterInitialized(cluster.ID)
}

func (lc *LiveCluster) GetClosestCluster(edgeNode *EdgeNode, checkResources bool) (*FogNode, float32) {
	var closestActive *FogNode
	var closestPing float32

	idx := 0
	for idx < len(edgeNode.SortedPings) && closestActive == nil {
		fn := lc.FogNodes[edgeNode.SortedPings[idx]]
		if fn.Active {
			closestActive = fn
			closestPing = edgeNode.Pings[idx]
			if checkResources && closestActive.IsFull() {
				closestActive = nil
			}
		}
		idx++
	}

	return closestActive, closestPing
}

func (lc *LiveCluster) GetClosestFogNode(edgeNode *EdgeNode, checkResources bool) (*FogNode, float32) {
	var closestFog *FogNode
	var closestPing float32

	if checkResources {
		idx := 0

		for idx < len(edgeNode.SortedPings) && closestFog == nil {
			closestFog = lc.FogNodes[edgeNode.SortedPings[idx]]
			closestPing = edgeNode.Pings[idx]
			if checkResources && closestFog.IsFull() {
				closestFog = nil
			}
			idx++
		}
	} else {
		closestFog = lc.FogNodes[edgeNode.SortedPings[0]]
	}

	return closestFog, closestPing
}

func (lc *LiveCluster) GetClosestClusterExcept(edgeNode *EdgeNode, except *FogNode, checkResources bool) (*FogNode, float32) {
	var closestActive *FogNode
	var closestPing float32

	idx := 0
	for idx < len(edgeNode.SortedPings) && (closestActive == nil || closestActive == except) {
		fn := lc.FogNodes[edgeNode.SortedPings[idx]]
		if fn.Active {
			closestActive = fn
			closestPing = edgeNode.Pings[idx]
			if checkResources && closestActive.IsFull() {
				closestActive = nil
			}
		}
		idx++
	}

	return closestActive, closestPing
}
