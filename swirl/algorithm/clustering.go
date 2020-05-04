package algorithm

import (
	"math/rand"
)

//NodeNames may look strange here, but it's required to make some things O(1) and doesn't use all that much memory.
type LiveCluster struct {
	NodeNames              map[string]*EdgeNode
	ClusteredNodes         map[*EdgeNode]*FogNode
	ActiveClusters         map[*FogNode]map[string]*EdgeNode //[]*EdgeNode
	FogNodes               []*FogNode
	ClusterCreationAllowed bool
}

func (lc *LiveCluster) Init() *LiveCluster {
	lc.ClusterCreationAllowed = true
	lc.NodeNames = make(map[string]*EdgeNode)
	lc.ClusteredNodes = make(map[*EdgeNode]*FogNode)
	lc.ActiveClusters = make(map[*FogNode]map[string]*EdgeNode) //[]*EdgeNode)

	return lc
}

func (lc *LiveCluster) InitStaticClusters(num int) {
	lc.NodeNames = make(map[string]*EdgeNode)
	lc.ClusteredNodes = make(map[*EdgeNode]*FogNode)
	lc.ActiveClusters = make(map[*FogNode]map[string]*EdgeNode)
	lc.ClusterCreationAllowed = false

	for _, node := range lc.FogNodes {
		node.Active = false
		node.ResourcesUsed = make(map[Resource]int)
	}

	for i := 0; i < num; i++ {
		found := false
		for !found {
			idx := rand.Int31n(int32(len(lc.FogNodes)))
			node := lc.FogNodes[idx]
			_, clusterExists := lc.ActiveClusters[node]
			if !clusterExists {
				lc.InitCluster(node)
				found = true
			}
		}
	}
}

func (lc *LiveCluster) ProcessNode(slaMaxPing float32, edgeNode *EdgeNode, checkResources bool) {
	existing, containsNode := lc.NodeNames[edgeNode.Name]
	if containsNode {
		lc.UpdateClusteredNode(slaMaxPing, existing, edgeNode, checkResources)
	} else {
		lc.AddNode(slaMaxPing, edgeNode, checkResources)
	}
}

func (lc *LiveCluster) AssignClosest(edgeNodes []*EdgeNode, checkResources bool) {
	/*existing, containsNode := lc.NodeNames[edgeNode.Name]
	if containsNode {
		lc.UpdateClusteredNode(slaMaxPing, existing, edgeNode, checkResources)
	} else {
		lc.AddNode(slaMaxPing, edgeNode, checkResources)
	}*/
	for _, fn := range lc.FogNodes {
		distances := make(map[float32]*EdgeNode)

		for _, en := range edgeNodes {
			distances[en.GetPing(fn)] = en
		}

		//actually, i now realize this test wouldn't do anything useful, so i'm giving up on it for now
		//eNodes, pings := SortENPings(distances)
	}
}

func (lc *LiveCluster) UpdateClusteredNode(slaMaxPing float32, existing *EdgeNode, newNode *EdgeNode, checkResources bool) {
	fogNode, _ := lc.ClusteredNodes[newNode]
	oldPing := existing.GetPing(fogNode) //.Pings[fogNode]
	newPing := newNode.GetPing(fogNode)  //.Pings[fogNode]

	existing.Pings = newNode.Pings
	existing.SortedPings = newNode.SortedPings

	//have to do something about it if ping went over max SLA
	if oldPing < slaMaxPing && newPing > slaMaxPing {
		lc.RemoveNode(slaMaxPing, existing, checkResources)
		lc.AddNode(slaMaxPing, existing, checkResources)
	}
}

func (lc *LiveCluster) RemoveNode(slaMaxPing float32, edgeNode *EdgeNode, checkResources bool) {
	fogNode := lc.ClusteredNodes[edgeNode]
	lc.RemoveFromCluster(edgeNode, fogNode)

	sameCluster := lc.ActiveClusters[fogNode]
	if (checkResources && fogNode.IsUnderutilized()) || (!checkResources && len(sameCluster) <= 200) {
		//try to move the remaining edge nodes
		failed := false
		newClusters := make(map[*EdgeNode]*FogNode)

		for len(sameCluster) > 0 && !failed {
			_, node := GetFirstItem(sameCluster)
			closestOther, otherPing := lc.GetClosestClusterExcept(node, fogNode, checkResources)
			if closestOther != nil {
				fPing := node.GetPing(fogNode)
				if fPing > slaMaxPing || (fPing < slaMaxPing && otherPing < slaMaxPing) {
					lc.RemoveFromCluster(node, fogNode)
					lc.AddToCluster(node, closestOther)
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
				lc.AddToCluster(node, fogNode)
			}
		} else {
			fogNode.Active = false
			delete(lc.ActiveClusters, fogNode)
			fogNode.ResourcesUsed = make(map[Resource]int)
		}
	}
}

func GetFirstItem(nodes map[string]*EdgeNode) (string, *EdgeNode) {
	for a, b := range nodes {
		return a, b
	}
	return "", nil
}

func (lc *LiveCluster) AddNode(slaMaxPing float32, edgeNode *EdgeNode, checkResources bool) {
	if len(lc.ActiveClusters) == 0 {
		lc.CreateNewClusterFor(edgeNode, nil, checkResources)
	} else {
		closestCluster, ping := lc.GetClosestCluster(edgeNode, checkResources)
		if closestCluster == nil && lc.ClusterCreationAllowed {
			lc.CreateNewClusterFor(edgeNode, nil, checkResources)
		} else if ping > slaMaxPing && lc.ClusterCreationAllowed {
			lc.TryCreateNewClusterFor(edgeNode, closestCluster, checkResources)
		} else {
			lc.AddToCluster(edgeNode, closestCluster)
		}
	}
}

func (lc *LiveCluster) TryCreateNewClusterFor(edgeNode *EdgeNode, closestCluster *FogNode, checkResources bool) {
	closestFogNode, _ := lc.GetClosestFogNode(edgeNode, checkResources)
	if closestFogNode == closestCluster {
		lc.AddToCluster(edgeNode, closestCluster)
	} else {
		lc.CreateNewClusterFor(edgeNode, closestFogNode, checkResources)
	}
}

func (lc *LiveCluster) CreateNewClusterFor(edgeNode *EdgeNode, fogNode *FogNode, checkResources bool) {
	if fogNode == nil {
		fogNode, _ = lc.GetClosestFogNode(edgeNode, checkResources)
	}

	_, exists := lc.ActiveClusters[fogNode]
	if exists {
		lc.AddToCluster(edgeNode, fogNode)
	} else {
		lc.InitCluster(fogNode)
		lc.AddToCluster(edgeNode, fogNode)
	}
}

func (lc *LiveCluster) AddToCluster(edgeNode *EdgeNode, cluster *FogNode) {
	lc.ClusteredNodes[edgeNode] = cluster
	lc.ActiveClusters[cluster][edgeNode.Name] = edgeNode
	lc.NodeNames[edgeNode.Name] = edgeNode

	cluster.ResourcesUsed[CPUShares] += 2
	cluster.ResourcesUsed[Memory] += 2
	cluster.ResourcesUsed[Network] += 1000
}

func (lc *LiveCluster) RemoveFromCluster(edgeNode *EdgeNode, cluster *FogNode) {
	delete(lc.ClusteredNodes, edgeNode)
	delete(lc.ActiveClusters[cluster], edgeNode.Name)
	delete(lc.NodeNames, edgeNode.Name)

	cluster.ResourcesUsed[CPUShares] -= 2
	cluster.ResourcesUsed[Memory] -= 2
	cluster.ResourcesUsed[Network] -= 1000
}

func (lc *LiveCluster) InitCluster(cluster *FogNode) {
	cluster.Active = true
	lc.ActiveClusters[cluster] = make(map[string]*EdgeNode)

	cluster.ResourcesUsed[CPUShares] = 200
	cluster.ResourcesUsed[Memory] = 200
	cluster.ResourcesUsed[Network] = 0
}

func (lc *LiveCluster) GetClosestCluster(edgeNode *EdgeNode, checkResources bool) (*FogNode, float32) {
	var closestActive *FogNode
	var closestPing float32

	idx := 0
	for idx < len(edgeNode.SortedPings) && closestActive == nil {
		if edgeNode.SortedPings[idx].Active {
			closestActive = edgeNode.SortedPings[idx]
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
			/*if !edgeNode.SortedPings[idx].IsFull() {
				closestFog = edgeNode.SortedPings[idx]
				closestPing = edgeNode.Pings[idx]
			}*/
			closestFog = edgeNode.SortedPings[idx]
			closestPing = edgeNode.Pings[idx]
			if checkResources && closestFog.IsFull() {
				closestFog = nil
			}
			idx++
		}
	} else {
		closestFog = edgeNode.SortedPings[0]
	}

	return closestFog, closestPing
}

func (lc *LiveCluster) GetClosestClusterExcept(edgeNode *EdgeNode, except *FogNode, checkResources bool) (*FogNode, float32) {
	var closestActive *FogNode
	var closestPing float32

	idx := 0
	for idx < len(edgeNode.SortedPings) && (closestActive == nil || closestActive == except) {
		if edgeNode.SortedPings[idx].Active {
			closestActive = edgeNode.SortedPings[idx]
			closestPing = edgeNode.Pings[idx]
			if checkResources && closestActive.IsFull() {
				closestActive = nil
			}
		}
		idx++
	}

	return closestActive, closestPing
}
