package main

import (
	"fmt"
	//corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"os"
	"swirl/swirlservice/algorithm"
	"swirl/swirlservice/config"
	"swirl/swirlservice/kube"
	"swirl/swirlservice/ws"
	//"k8s.io/apimachinery/pkg/fields"
	//"encoding/json"
	//v1 "k8s.io/api/apps/v1"
)

func main() {
	argsWithoutProg := os.Args[1:]
	cfgFile := "defaultconfig.json"
	if len(argsWithoutProg) > 0 {
		cfgFile = argsWithoutProg[0]
	}

	config.LoadConfig(cfgFile)

	/*mlabels := make(map[string]string)
	mlabels["k8s-app"] = "testservice"
	matchLabels := make(map[string]string)
	matchLabels["name"] = "testservice"
	//matchLabels["app"] = "edgetest"
	nodeLabels := make(map[string]string)
	nodeLabels["swirlnodetype"] = "edgenode"
	container := corev1.Container{
		Name:    "testservice",
		Image:   "togoetha/go-rest",
		Command: []string{"/app/hello"},
	}
	dep := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testservice",
			Namespace: "default",
			Labels:    mlabels,
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabels,
				},
				Spec: corev1.PodSpec{
					//HostNetwork:  true,
					//NodeSelector: nodeLabels,
					Containers: []corev1.Container{container},
				},
			},
		},
	}
	jsonbytes, _ := json.Marshal(dep)
	fmt.Println(string(jsonbytes))*/

	algorithm.Init()
	kube.Init()
	//go func() { }()

	router := ws.NewRouter()
	port := fmt.Sprintf(":%d", config.Cfg.Port)
	log.Fatal(http.ListenAndServe(port, router))
}
