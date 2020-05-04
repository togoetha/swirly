package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	//"github.com/gorilla/mux"
)

//GET /setFogNodes
func SetFogNodes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Ping")

	pingIPs := []string{}
	err := json.NewDecoder(r.Body).Decode(&pingIPs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updateFogAddresses(pingIPs)
}
