package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"swirl/swirlservice/algorithm"
	"swirl/swirlservice/config"
	//"swirl/swirlservice/ws"
	//"github.com/gorilla/mux"
)

//POST /updateFogNodePings
func UpdateFogNodePings(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateFogNodePings")

	fogpings := PingUpdate{}
	err := json.NewDecoder(r.Body).Decode(&fogpings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pings := []float32{}
	fogids := []string{}
	for i := 0; i < len(fogpings.Pings); i++ {
		pings = append(pings, fogpings.Pings[i].Ping)
		fogids = append(fogids, fogpings.Pings[i].Node)
	}

	node := &algorithm.EdgeNode{
		ID:          fogpings.NodeId,
		Pings:       pings,
		SortedPings: fogids,
	}
	algorithm.Clusterer.ProcessEdgeNode(config.Cfg.MaxPing, node, true)
}

func UpdateFogNodeResources(w http.ResponseWriter, r *http.Request) {
	fmt.Println("UpdateFogNodeResources")

	resUpdate := ResourceUpdate{}

	err := json.NewDecoder(r.Body).Decode(&resUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("For %s\n", resUpdate.NodeId)

	changed := algorithm.Clusterer.ProcessFogNode(resUpdate.NodeId, resUpdate.Resources, resUpdate.TotalResources)
	fmt.Printf("Changed %t\n", changed)
	if changed {
		UpdateFogNodeLists()
	}
}

func GetFogNodeIPs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetFogNodeIPs")

	fogIPs := getFogIPs()

	json, err := json.Marshal(fogIPs)
	fmt.Printf("Responding %s\n", string(json))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(json)
}
