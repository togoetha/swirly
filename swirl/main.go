package main

import (
	"fmt"
	"math"
	"os"
	"swirl/swirl/algorithm"
	"swirl/swirl/config"
)

func main() {
	argsWithoutProg := os.Args[1:]
	cfgFile := "defaultconfig.json"
	if len(argsWithoutProg) > 0 {
		cfgFile = argsWithoutProg[0]
	}

	//fmt.Printf("Loading config file %s\n", cfgFile)
	config.LoadConfig(cfgFile)

	//determine what type of test needs to be done. It's best to do only one of these at a time, but in theory they can all run in sequence.
	if config.Cfg.SpeedTest {
		DoSpeedSimulation()
	}
	if config.Cfg.MemTest {
		DoMemSimulation()
	}
	if config.Cfg.ClusterTest {
		DoClusterStatsTest()
	}
}

//Rounds a number up to a multiple of FogNodeStep, required for finding the minimum number of fog nodes for an edge infrastructure and then
//making it adhere to the test settings
func RoundUp(number float64) int {
	steps := number / float64(config.Cfg.FogNodeStep)
	return config.Cfg.FogNodeStep * int(math.Ceil(steps))
}

//This builds a service topology and measures the timings of the add and delete operations.
//For automation purposes, it iterates from MinEdgeNodes to MaxEdgeNodes in EdgeNodeStep steps.
//The same goes for MinFogNodes, MaxFogNodes and FogNodeStep, however MinFogNodes can be overriden by the magic number 800 (see below).
//Because the generated fog nodes and edge nodes are completely random, any topology may turn out to be wildly positive or negative, skewing the results.
//Therefore, it is recommended to do a good number of runs per edgenode/fognode step (Iterations = 20 seems good).
//TODO: get rid of the magic number "800": it has to do with how many clients can "safely" fit on the average fog node given the hardcoded
//resource constraints in clustering.go.
//In other words, if we take more than 800, the algorithm will GIGO because it can't find a spot for every edge node, even a bad one.
func DoSpeedSimulation() {
	clusterStats := []algorithm.GroupStats{}
	deleteStats := []algorithm.GroupStats{}

	printHeader := true
	checkResources := config.Cfg.CheckResources
	slaMaxPing := float32(config.Cfg.SLAMaxPing)
	maxPingDiff := float64(config.Cfg.MaxPingDiff)
	deleteAmount := config.Cfg.AmountDeleteNodes

	//iterate over number of edge nodes
	for en := config.Cfg.MinEdgeNodes; en <= config.Cfg.MaxEdgeNodes; en += config.Cfg.EdgeNodeStep {

		//iterate over number of fog nodes, start at a minimum of 1 per 800 edge nodes
		fnLimit := float64(RoundUp(float64(en) / 800))
		for fn := int(math.Max(fnLimit, float64(config.Cfg.MinFogNodes))); fn <= config.Cfg.MaxFogNodes; fn += config.Cfg.FogNodeStep {
			clusterLines := []algorithm.StatsLine{}
			deleteLines := []algorithm.StatsLine{}

			for iter := 0; iter < config.Cfg.Iterations; iter++ {
				algorithm.GenerateNodes(en, fn, maxPingDiff)
				clStats := algorithm.ClusterIncrementalTimeStats(slaMaxPing, checkResources)
				clStats.EdgeNodes = en
				clusterLines = append(clusterLines, clStats)

				delStats := algorithm.DeleteNodes(slaMaxPing, deleteAmount, checkResources)
				delStats.EdgeNodes = en
				deleteLines = append(deleteLines, delStats)
			}

			//print out stats and make the group lines
			for _, l := range clusterLines {
				line, _ := l.String()
				if printHeader {
					fmt.Println(l.LineHeader())
					printHeader = false
				}
				fmt.Println(line)
			}
			fmt.Println()
			cGroupStat := algorithm.MakeGroupStats(clusterLines)
			clusterStats = append(clusterStats, cGroupStat)
			for _, l := range deleteLines {
				line, _ := l.String()
				//fmt.Println(header)
				fmt.Println(line)
			}
			fmt.Println()
			dGroupStat := algorithm.MakeGroupStats(deleteLines)
			deleteStats = append(deleteStats, dGroupStat)
		}

	}

	fmt.Println(clusterStats[0].GroupHeader())
	//fmt.Println(algorithm.GroupHeader())
	for _, l := range clusterStats {
		line, _ := l.String()
		fmt.Println(line)
	}
	for _, l := range deleteStats {
		line, _ := l.String()
		fmt.Println(line)
	}
}

//Very similar to speed test, but this one only constructs the service topology and measures the number of used fog nodes, in addition
// to some distance stats.
//Distance stats include AvgMinDist (100% ideal, impossible when resource constraints are enabled), AvgDist (algorithm result), and AvgRndDist
//(result if the same number of services were randomly deployed across the service topology).
func DoClusterStatsTest() {
	clusterStats := []algorithm.GroupStats{}

	printHeader := true
	checkResources := config.Cfg.CheckResources
	slaMaxPing := float32(config.Cfg.SLAMaxPing)
	maxPingDiff := float64(config.Cfg.MaxPingDiff)

	//iterate over number of edge nodes
	for en := config.Cfg.MinEdgeNodes; en <= config.Cfg.MaxEdgeNodes; en += config.Cfg.EdgeNodeStep {

		//iterate over number of fog nodes
		fnLimit := float64(RoundUp(float64(en) / 800))
		for fn := int(math.Max(fnLimit, float64(config.Cfg.MinFogNodes))); fn <= config.Cfg.MaxFogNodes; fn += config.Cfg.FogNodeStep {
			clusterLines := []algorithm.StatsLine{}

			for iter := 0; iter < config.Cfg.Iterations; iter++ {
				algorithm.GenerateNodes(en, fn, maxPingDiff)
				clStats := algorithm.ClusterIncrementalClusterStats(slaMaxPing, checkResources)
				clStats.EdgeNodes = en

				rndStats := algorithm.ClusterIncrementalStaticFognodes(slaMaxPing, checkResources, 0)
				clStats.Metrics["AvgRndDist"] = rndStats.Metrics["AvgDist"]
				rndStats := algorithm.ClosestClusterIncrementalStaticFognodes(slaMaxPing, checkResources, 0)
				clStats.Metrics["AvgCRndDist"] = rndStats.Metrics["AvgDist"]
				clStats.MNames = []string{"AvgDist", "AvgMinDist", "AvgRndDist", "AvgCRndDist"}
				clusterLines = append(clusterLines, clStats)
			}

			//print out stats and make the group lines
			for _, l := range clusterLines {
				if printHeader {
					fmt.Println(l.LineHeader())
					printHeader = false
				}
				line, _ := l.String()
				fmt.Println(line)
			}
			fmt.Println()
			cGroupStat := algorithm.MakeGroupStats(clusterLines)
			clusterStats = append(clusterStats, cGroupStat)
		}

	}

	fmt.Println(clusterStats[0].GroupHeader())
	for _, l := range clusterStats {
		line, _ := l.String()
		//fmt.Println(header)
		fmt.Println(line)
	}

}

//Builds a service topology and prints the memory use of the process.
//Note: this can't run more than once because the golang GC doesn't work well enough and the memory use stays almost constant across runs,
// all loops are taken care of in a shell script.
func DoMemSimulation() {
	checkResources := config.Cfg.CheckResources
	slaMaxPing := float32(config.Cfg.SLAMaxPing)
	maxPingDiff := float64(config.Cfg.MaxPingDiff)

	en := config.Cfg.MinEdgeNodes

	fnLimit := float64(RoundUp(float64(en) / 800))
	fn := int(math.Max(fnLimit, float64(config.Cfg.MinFogNodes)))

	clStats := algorithm.GetMemoryForClusters(en, fn, maxPingDiff, slaMaxPing, checkResources)
	clStats.EdgeNodes = en
	fmt.Println(clStats.String())
}
