package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Cfg *Config

type Config struct {
	Port                  int     `json:"port"`
	FogNodeLabel          string  `json:"fogNodeLabel"`
	EdgeNodeLabel         string  `json:"edgeNodeLabel"`
	EdgeDeploymentName    string  `json:"edgeDeploymentName"`
	FogDeploymentTemplate string  `json:"fogDeploymentTemplate"`
	MaxPing               float32 `json:"maxPing"`
	FogPort               int     `json:"fogPort"`
	EdgePort              int     `json:"edgePort"`
	EdgeUpdateURL         string  `json:"edgeUpdateUrl"`
}

func LoadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		//return err
	}
	decoder := json.NewDecoder(file)
	Cfg = &Config{}
	err = decoder.Decode(Cfg)
	if err != nil {
		fmt.Println(err.Error())
		//return err
	}

	if os.Getenv("EDGE_DEPLOYMENT_NAME") != "" {
		Cfg.EdgeDeploymentName = os.Getenv("EDGE_DEPLOYMENT_NAME")
	}

	return err
}
