package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"swirl/edgeservice/config"
	"time"
)

var fogNodePings map[string]float32

//var nodeID string

func startFogNodePings() {
	fogNodePings = make(map[string]float32)
	fetchFogNodeIPs()
	for true {
		updateFogPings()
		time.Sleep(config.Cfg.PingPeriod * time.Second)
	}
}

func fetchFogNodeIPs() {
	fullURL := fmt.Sprintf("http://%s:%d/%s", config.Cfg.SwirlServer, config.Cfg.SwirlPort, config.Cfg.FetchFogURL)
	fmt.Printf("Fetching fog pings %s\n", fullURL)
	response, err := http.Get(fullURL)

	if err != nil {
		fmt.Println(err.Error())
	}

	fogNodeIPs := []string{}
	err = json.NewDecoder(response.Body).Decode(&fogNodeIPs)
	if err != nil {
		return
	}

	for _, fogNodeIP := range fogNodeIPs {
		fogNodePings[fogNodeIP] = -1
	}
}

func updateFogPings() {
	reportChanges := false

	for ip, oldPing := range fogNodePings {
		newPing := getPing(ip)

		ratio := newPing / oldPing
		if ratio > 1.3 || ratio < 0.7 || (newPing > config.Cfg.MaxPing && oldPing < config.Cfg.MaxPing) {
			reportChanges = true
		}
		fogNodePings[ip] = newPing
	}

	if reportChanges {
		reportPingChanges()
	}
}

func getPing(ip string) float32 {
	start := time.Now()

	fullURL := fmt.Sprintf("http://%s:%d/%s", ip, config.Cfg.PingPort, config.Cfg.PingURL)
	fmt.Printf("Pinging %s\n", fullURL)
	response, err := http.Get(fullURL)

	stop := time.Now()

	newPing := float32(stop.Sub(start).Nanoseconds()) / 1000000
	if response.StatusCode != 200 || err != nil {
		fmt.Println(err.Error())
		newPing = -1
	}
	return newPing
}

type PingUpdate struct {
	NodeId string
	Pings  []NodePing
}

func reportPingChanges() {
	sortedPings := SortNodePings(fogNodePings)
	update := PingUpdate{
		NodeId: config.Cfg.NodeID,
		Pings:  sortedPings,
	}

	resJson, err := json.Marshal(update)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fullURL := fmt.Sprintf("http://%s:%d/%s", config.Cfg.SwirlServer, config.Cfg.SwirlPort, config.Cfg.PingReportURL)
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

func updateFogAddresses(ips []string) {
	for oldip, _ := range fogNodePings {
		stillPresent := false
		for _, newip := range ips {
			if newip == oldip {
				stillPresent = true
			}
		}
		if !stillPresent {
			delete(fogNodePings, oldip)
		}
	}
	for _, ip := range ips {
		_, found := fogNodePings[ip]
		if !found {
			fogNodePings[ip] = -1
		}
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//Again, a horrible golang mess. Because golang provides no real way to sort maps according to their key values,
//this thing had to be written.
func SortNodePings(pings map[string]float32) []NodePing {
	nodePings := []NodePing{}
	for fn, ping := range pings {
		nPing := NodePing{Node: fn, Ping: ping}
		nodePings = append(nodePings, nPing)
	}

	ping := func(p1, p2 *NodePing) bool {
		return p1.Ping < p2.Ping
	}
	By(ping).Sort(nodePings)

	/*nodes := []string{}
	npings := []float32{}
	for i := 0; i < len(nodePings); i++ {
		nodes = append(nodes, nodePings[i].Node)
		npings = append(npings, nodePings[i].Ping)
	}
	return nodes, npings*/
	return nodePings
}

type NodePing struct {
	Node string
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
