package algorithm

import (
	"fmt"
	"sort"
)

//I can explain this mess of a file! golang's range keyword (for looping over maps and arrays) returns indexes in a randomized way.
//Therefore, each map has an array with all its string indexes that can be iterated over sequentially to keep their order.
//It's the only way to output the map values in the same order for each StatsLine.
//By the way, a stats line represents the statistics of a single constructed topology, but its fields are variable (time, memory, cluster sizes..)
//Metrics are important ones, they get more statistical processing in GroupStats, for Stats only the median gets selected.
type StatsLine struct {
	EdgeNodes int
	FogNodes  int

	Metrics map[string]float64
	Stats   map[string]float64

	MNames []string
	SNames []string
}

//Sort of constructor
func (sl StatsLine) Init() StatsLine {
	sl.Metrics = make(map[string]float64)
	sl.Stats = make(map[string]float64)
	sl.SNames = []string{}
	sl.MNames = []string{}
	return sl
}

//Create a header with all the field names
func (sl *StatsLine) LineHeader() string {
	var line = "ENodes;FNodes"
	for _, name := range sl.SNames {
		line += fmt.Sprintf(";%s", "Med"+name)
	}
	for _, name := range sl.MNames {
		line += ";" + name
	}
	return line
}

//Create a string (+header) with all the values
func (sl *StatsLine) String() (string, string) {
	header := "ENodes;FNodes"
	line := fmt.Sprintf("%d;%d", sl.EdgeNodes, sl.FogNodes)

	for _, name := range sl.SNames {
		header += ";" + name
		line += fmt.Sprintf(";%f", sl.Stats[name])
	}
	for _, name := range sl.MNames {
		header += ";" + name
		line += fmt.Sprintf(";%f", sl.Metrics[name])
	}
	return line, header
}

//GroupsStats is used to calculate min, median and max of every metric contained in a StatsLine, and the median of every "stat".
type GroupStats struct {
	EdgeNodes int
	FogNodes  int

	MinMetrics map[string]float64
	MedMetrics map[string]float64
	MaxMetrics map[string]float64
	MedStats   map[string]float64

	MNames []string
	SNames []string
}

func MakeGroupStats(lines []StatsLine) GroupStats {
	gs := GroupStats{
		EdgeNodes:  lines[0].EdgeNodes,
		FogNodes:   lines[0].FogNodes,
		MinMetrics: make(map[string]float64),
		MedMetrics: make(map[string]float64),
		MaxMetrics: make(map[string]float64),
		MedStats:   make(map[string]float64),
		MNames:     []string{},
		SNames:     []string{},
	}

	for i := 0; i < len(lines[0].MNames); i++ {
		name := lines[0].MNames[i]
		vals := []float64{}
		gs.MNames = append(gs.MNames, name)
		for _, line := range lines {
			vals = append(vals, line.Metrics[name])
		}

		gs.MinMetrics[name] = Min(vals)
		gs.MedMetrics[name] = Med(vals)
		gs.MaxMetrics[name] = Max(vals)
	}

	for i := 0; i < len(lines[0].SNames); i++ {
		name := lines[0].SNames[i]
		vals := []float64{}
		gs.SNames = append(gs.SNames, name)
		for _, line := range lines {
			vals = append(vals, line.Stats[name])
		}
		gs.MedStats[name] = Med(vals)
	}
	return gs
}

func Min(vals []float64) float64 {
	min := vals[0]

	for idx := 0; idx < len(vals); idx++ {
		if vals[idx] < min {
			min = vals[idx]
		}
	}
	return min
}

func Max(vals []float64) float64 {
	max := vals[0]

	for idx := 0; idx < len(vals); idx++ {
		if vals[idx] > max {
			max = vals[idx]
		}
	}
	return max
}

func Med(vals []float64) float64 {
	sort.Float64s(vals)

	return vals[len(vals)/2]
}

func (gs GroupStats) GroupHeader() string {
	line := "ENodes;FNodes"
	for _, name := range gs.SNames {
		line += fmt.Sprintf(";%s", "Med"+name)
	}
	for _, name := range gs.MNames {
		line += fmt.Sprintf(";%s;%s;%s", "Min"+name, "Med"+name, "Max"+name)
	}
	return line
}

func (sl GroupStats) String() (string, string) {
	header := "ENodes;FNodes"
	line := fmt.Sprintf("%d;%d", sl.EdgeNodes, sl.FogNodes)

	for i := 0; i < len(sl.SNames); i++ {
		name := sl.SNames[i]
		line += fmt.Sprintf(";%f", sl.MedStats[name])
		header += fmt.Sprintf(";%s", name)
	}

	for i := 0; i < len(sl.MNames); i++ {
		name := sl.MNames[i]
		line += fmt.Sprintf(";%f;%f;%f", sl.MinMetrics[name], sl.MedMetrics[name], sl.MaxMetrics[name])
		header += fmt.Sprintf(";%s;%s;%s", "Min"+name, "Med"+name, "Max"+name)
	}
	return line, header
}
