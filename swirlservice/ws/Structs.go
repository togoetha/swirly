package ws

import (
	"swirl/swirlservice/algorithm"
)

type NodePing struct {
	Node string
	Ping float32
}

type PingUpdate struct {
	NodeId string
	Pings  []NodePing
}

type ResourceUpdate struct {
	NodeId         string
	Resources      map[algorithm.Resource]int
	TotalResources map[algorithm.Resource]int
}
