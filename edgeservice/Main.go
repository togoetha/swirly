package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"swirl/edgeservice/config"
)

var kubernetesHost string
var kubernetesPort string
var defaultPodFile string

func main() {
	argsWithoutProg := os.Args[1:]
	cfgFile := "defaultconfig.json"
	if len(argsWithoutProg) > 0 {
		cfgFile = argsWithoutProg[0]
	}

	config.LoadConfig(cfgFile)

	go func() { startFogNodePings() }()

	router := NewRouter()
	port := fmt.Sprintf(":%d", config.Cfg.Port)
	log.Fatal(http.ListenAndServe(port, router))
}
