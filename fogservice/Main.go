package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"swirl/fogservice/config"
)

var kubernetesHost string
var kubernetesPort string
var defaultPodFile string

var pings map[string]map[string]int

func main() {
	argsWithoutProg := os.Args[1:]
	cfgFile := "defaultconfig.json"
	if len(argsWithoutProg) > 0 {
		cfgFile = argsWithoutProg[0]
	}

	//this is just to simulate latencies in the k8s cluster, remove in any other situation
	pings = make(map[string]map[string]int)
	//fog1
	pings["10.2.33.14"] = make(map[string]int)
	pings["10.2.33.14"]["10.2.33.39"] = 30
	pings["10.2.33.14"]["10.2.33.8"] = 50

	pings["10.2.33.14"]["10.2.33.17"] = 110
	pings["10.2.33.14"]["10.2.33.10"] = 130

	pings["10.2.33.14"]["10.2.33.9"] = 55
	pings["10.2.33.14"]["10.2.33.41"] = 120
	//fog2
	pings["10.2.33.38"] = make(map[string]int)
	pings["10.2.33.38"]["10.2.33.39"] = 140
	pings["10.2.33.38"]["10.2.33.8"] = 120

	pings["10.2.33.38"]["10.2.33.17"] = 20
	pings["10.2.33.38"]["10.2.33.10"] = 60

	pings["10.2.33.38"]["10.2.33.9"] = 130
	pings["10.2.33.38"]["10.2.33.41"] = 65

	//fog3
	pings["10.2.33.32"] = make(map[string]int)
	pings["10.2.33.32"]["10.2.33.39"] = 95
	pings["10.2.33.32"]["10.2.33.8"] = 85

	pings["10.2.33.32"]["10.2.33.17"] = 120
	pings["10.2.33.32"]["10.2.33.10"] = 110

	pings["10.2.33.32"]["10.2.33.9"] = 40
	pings["10.2.33.32"]["10.2.33.41"] = 55

	ootput, _ := json.Marshal(pings)
	fmt.Println(ootput)

	config.LoadConfig(cfgFile)

	go func() { startUpdateResources() }()

	router := NewRouter()
	port := fmt.Sprintf(":%d", config.Cfg.Port)
	log.Fatal(http.ListenAndServe(port, router))
}
