package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Cfg *Config

type Config struct {
	MinEdgeNodes      int  `json:"minEdgeNodes"`
	MaxEdgeNodes      int  `json:"maxEdgeNodes"`
	EdgeNodeStep      int  `json:"edgeNodeStep"`
	MinFogNodes       int  `json:"minFogNodes"`
	MaxFogNodes       int  `json:"maxFogNodes"`
	FogNodeStep       int  `json:"fogNodeStep"`
	Iterations        int  `json:"iterations"`
	CheckResources    bool `json:"checkResources"`
	MaxPingDiff       int  `json:"maxPingDiff"`
	SLAMaxPing        int  `json:"slaMaxPing"`
	AmountDeleteNodes int  `json:"amountDeleteNodes"`
	SpeedTest         bool `json:"speedTest"`
	MemTest           bool `json:"memTest"`
	ClusterTest       bool `json:"clusterTest"`
}

func LoadConfig(filename string) error {
	//fmt.Printf("Loading config %s\n", filename)
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

	/*fmt.Printf("VkubeServiceURL check %s\n", Cfg.VkubeServiceURL)
	if Cfg.VkubeServiceURL == "" {
		fmt.Printf("Loading from env instead")
		Cfg.Runtime = os.Getenv("FLEDGE_RUNTIME")
		Cfg.DeviceName = os.Getenv("FLEDGE_DEVICE_NAME")
		Cfg.DeviceIP = os.Getenv("FLEDGE_DEVICE_IP")
		Cfg.ServicePort = os.Getenv("FLEDGE_SERVICE_PORT")
		Cfg.KubeletPort = os.Getenv("FLEDGE_KUBELET_PORT")
		Cfg.VkubeServiceURL = os.Getenv("FLEDGE_VKUBE_URL")
		Cfg.IgnoreKubeProxy = os.Getenv("FLEDGE_IGNORE_KPROXY")
		Cfg.Interface = os.Getenv("FLEDGE_INET_INTERFACE")
	}*/

	return err
}
