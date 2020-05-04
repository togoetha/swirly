package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"swirl/fogservice/config"
	"time"
)

type ResourceUpdate struct {
	NodeId         string
	Resources      map[Resource]int
	TotalResources map[Resource]int
}

var cores int

//var nodeID string

func startUpdateResources() {
	cores = getCores()
	for true {
		updateResources()
		time.Sleep(config.Cfg.ResourceUpdatePeriod * time.Second)
	}
}

type Resource string

const (
	CPUShares Resource = "cpushares"
	Memory    Resource = "memory"
	Disk      Resource = "disk"
	Network   Resource = "network"
)

func updateResources() {
	resources, totalResources := getResources()

	resUpdate := ResourceUpdate{
		NodeId:         config.Cfg.NodeID,
		Resources:      resources,
		TotalResources: totalResources,
	}

	resJson, err := json.Marshal(resUpdate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fullURL := fmt.Sprintf("http://%s:%d/%s", config.Cfg.SwirlServer, config.Cfg.SwirlPort, config.Cfg.ResourceReportURL)
	fmt.Printf("Calling %s\n", fullURL)
	response, err := http.Post(fullURL, "application/json", bytes.NewBuffer(resJson))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	_, err = ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if err != nil {
		fmt.Println(err.Error())
	}
}

func getResources() (map[Resource]int, map[Resource]int) {
	resources := make(map[Resource]int)

	resources[CPUShares] = getCPU()
	resources[Memory] = getMemory()

	totalResources := make(map[Resource]int)

	totalResources[CPUShares] = 1000 * cores
	totalResources[Memory] = getTotalMemory()
	return resources, totalResources
}

func getTotalMemory() int {
	//Get memory
	memFree, _ := ExecCmdBash("free -m | grep 'Mem:'")
	//fmt.Println(memFree)
	memSize := strings.Split(reInsideWhtsp.ReplaceAllString(memFree, " "), " ")[1]
	memInt, _ := strconv.Atoi(memSize)
	return memInt
}

func getTotalStorage() string {
	//Get disk
	diskFree, _ := ExecCmdBash("df -h | grep -E '[[:space:]]/$'")
	//fmt.Println(diskFree)
	diskSize := strings.Split(reInsideWhtsp.ReplaceAllString(diskFree, " "), " ")[1]
	return diskSize
}

func getCores() int {
	stdout, _ := ExecCmdBash("nproc")
	numCpus := strings.Trim(string(stdout), "\n")
	//fmt.Println(numCpus)
	cpusInt, _ := strconv.Atoi(numCpus)
	return cpusInt
}

func getCPU() int {
	iostatc, _ := ExecCmdBash("iostat -c")
	var cpuUsed string
	var cpuLines = strings.Split(iostatc, "\n")
	//fmt.Println(memFree)
	cpuUsed = strings.Split(reInsideWhtsp.ReplaceAllString(cpuLines[3], " "), " ")[6]

	cpuPct, _ := strconv.ParseFloat(cpuUsed, 64)
	return int((100 - cpuPct) * float64(cores) * 10)
}

var reInsideWhtsp = regexp.MustCompile(`\s+`)

func getMemory() int {
	//Get memory
	//there's different types of output of the free command, trying the one with -/+ buffers/cache: first
	memFree, _ := ExecCmdBash("free -m | grep '-/+ buffers/cache:'")
	var memSize string
	if memFree != "" {
		//fmt.Println(memFree)
		memSize = strings.Split(reInsideWhtsp.ReplaceAllString(memFree, " "), " ")[2]
	} else {
		memFree, _ = ExecCmdBash("free -m | grep 'Mem:'")
		//fmt.Println(memFree)
		memSize = strings.Split(reInsideWhtsp.ReplaceAllString(memFree, " "), " ")[6]
	}
	memMb, _ := strconv.Atoi(memSize)
	return int(memMb)
}

func ExecCmdBash(dfCmd string) (string, error) {
	fmt.Printf("Executing %s\n", dfCmd)
	cmd := exec.Command("sh", "-c", dfCmd)
	stdout, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return "", err
	}
	//fmt.Println(string(stdout))
	return string(stdout), nil
}
