package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"swirl/fogservice/config"
	"time"
	//"github.com/gorilla/mux"
)

//GET /ping
func Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Ping")

	//remove when not simulating
	ip := strings.Split(r.RemoteAddr, ":")[0]
	ping := pings[config.Cfg.NodeID][ip]
	time.Sleep(time.Duration(ping) * time.Millisecond)

	w.Write([]byte("OK"))
	//w.WriteHeader(200)
}

//POST /setID
func GetResources(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetResources")

	resources, totalResources := getResources()

	resUpdate := ResourceUpdate{
		NodeId:         config.Cfg.NodeID,
		Resources:      resources,
		TotalResources: totalResources,
	}

	json, err := json.Marshal(resUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(json)
}
