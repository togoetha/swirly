package algorithm

import (
	"fmt"
	//"math"
)

type Resource string

const (
	CPUShares Resource = "cpushares"
	Memory    Resource = "memory"
	Disk      Resource = "disk"
	Network   Resource = "network"
)

type EdgeNode struct {
	ID string
	//X           float64
	//Y           float64
	SortedPings []string
	Pings       []float32
	//Pings       map[*FogNode]float32 //remove the map!!!
}

func (en *EdgeNode) Init() *EdgeNode {
	en.Pings = []float32{} //make(map[*FogNode]float32)
	return en
}

/*func (en *EdgeNode) Distance(other *FogNode) float64 {
	return math.Sqrt(math.Pow(en.X-other.X, 2) + math.Pow(en.Y-other.Y, 2))
}*/

func (en *EdgeNode) String() string {
	return fmt.Sprintf("ID %s", en.ID) //"X%d Y%d", en.X, en.Y)
}

type FogNode struct {
	ID string
	//X              float64
	//Y              float64
	Active         bool
	ResourcesUsed  map[Resource]int
	ResourcesTotal map[Resource]int
}

func (fn *FogNode) Init() *FogNode {
	fn.ResourcesTotal = make(map[Resource]int)
	fn.ResourcesUsed = make(map[Resource]int)
	return fn
}

func (fn *FogNode) String() string {
	return fmt.Sprintf("ID %s", fn.ID) //"X%d Y%d", fn.X, fn.Y)
}

func (fn *FogNode) IsFull() bool {
	full := false

	for res, amount := range fn.ResourcesUsed {
		fmt.Printf("Checking node %s resource %s, %d > %d\n", fn.ID, res, amount, fn.ResourcesTotal[res])
		if fn.ResourcesTotal[res] > 0 && amount >= fn.ResourcesTotal[res] {
			full = true
		}
	}

	return full
}

func (fn *FogNode) IsUnderutilized() bool {
	underfull := true

	for res, amount := range fn.ResourcesUsed {
		if float64(amount) >= 0.2*float64(fn.ResourcesTotal[res]) {
			underfull = false
		}
	}

	return underfull
}

func (fn *EdgeNode) GetPing(fNode *FogNode) float32 {
	idx := 0
	var ping float32
	found := false
	for idx < len(fn.SortedPings) && !found {
		if fn.SortedPings[idx] == fNode.ID {
			ping = fn.Pings[idx]
			found = true
		}

		idx++
	}
	return ping
}
