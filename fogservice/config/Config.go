package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var Cfg *Config

type Config struct {
	Port                 int           `json:"port"`
	SwirlServer          string        `json:"swirlServer"`
	SwirlPort            int           `json:"swirlPort"`
	ResourceReportURL    string        `json:"resourceReportUrl"`
	ResourceUpdatePeriod time.Duration `json:"resourceUpdatePeriod"`
	NodeID               string        `json:"nodeID"`
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

	fmt.Printf("NodeID check %s\n", Cfg.NodeID)
	if os.Getenv("NODEID") != "" {
		fmt.Printf("Loading nodeID from env instead")
		Cfg.NodeID = os.Getenv("NODEID")
	}
	fmt.Printf("SwirlServer check %s\n", Cfg.SwirlServer)
	if os.Getenv("SWIRLSERVER") != "" {
		fmt.Printf("Loading SwirlServer from env instead")
		Cfg.SwirlServer = os.Getenv("SWIRLSERVER")
	}

	return err
}
