package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"swirl/swirlservice/algorithm"
	"swirl/swirlservice/config"
)

func UpdateFogNodeLists() {
	fogIPs := getFogIPs()
	nodeJson, _ := json.Marshal(fogIPs)

	for _, node := range algorithm.Clusterer.EdgeNodes {
		fullURL := fmt.Sprintf("http://%s:%d/%s", node.ID, config.Cfg.EdgePort, config.Cfg.EdgeUpdateURL)
		fmt.Printf("Calling %s\n", fullURL)
		response, err := http.Post(fullURL, "application/json", bytes.NewBuffer(nodeJson))
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
}

func getFogIPs() []string {
	fogIPs := []string{}

	for _, node := range algorithm.Clusterer.FogNodes {
		fogIPs = append(fogIPs, node.ID)
	}
	return fogIPs
}
