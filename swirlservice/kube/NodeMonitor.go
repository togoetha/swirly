package kube

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/tools/cache"
	"encoding/json"
	"os"
	//"fmt"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"time"
	//"net/http"
	"swirl/swirlservice/algorithm"
	"swirl/swirlservice/config"
	"swirl/swirlservice/ws"
)

func addEdgeNode(nodekey interface{}) {
	//check if this works
	/*node := nodekey.(*corev1.Node)

	nodeID := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeHostName {
			nodeID = addr.Address
		}
	}
	algorithm.Clusterer.EdgeNodes[nodeID] = &algorithm.EdgeNode{
		ID:          nodeID,
		SortedPings: []string{},
		Pings:       []float32{},
	}*/
}

func addDeployment(podkey interface{}) {
	pod := podkey.(*corev1.Pod)
	nodeID := pod.Status.HostIP //pod.Spec.NodeName //.ObjectMeta.OwnerReferences[0].Name

	if nodeID == "" {
		fmt.Println("No IP assigned yet, waiting")

		for nodeID == "" {
			time.Sleep(time.Second)
			fmt.Println("Retrying deployment IP fetch")
			newpod, err := k8sClient.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Println(err.Error())
			} else {
				nodeID = newpod.Status.HostIP
			}
		}
	}
	fmt.Printf("Deployment detected %s\n", nodeID)
	algorithm.Clusterer.ProcessDeployment(config.Cfg.MaxPing, nodeID, true)

}

//A fog node was added through K8S
//There's a lot that can go wrong here, but it's fine for POC
func addFogNode(nodekey interface{}) {
	/*node := nodekey.(*corev1.Node)

	//First, get the nodeID (which is IP address for a fog node)
	nodeID := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeExternalIP {
			nodeID = addr.Address
		} else if addr.Type == v1.NodeInternalIP && nodeID == "" {
			nodeID = addr.Address
		}
	}

	//Get the initially free resources from the node
	fullURL := fmt.Sprintf("http://%s:%d/getResources", nodeID, config.Cfg.FogPort)
	fmt.Printf("Calling %s\n", fullURL)
	response, err := http.Get(fullURL)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	resUpdate := ws.ResourceUpdate{}
	err = json.NewDecoder(response.Body).Decode(&resUpdate)
	if err != nil {
		return
	}

	//And add it to the clusterer
	algorithm.Clusterer.ProcessFogNode(nodeID, resUpdate.Resources, resUpdate.TotalResources)*/
}

//An edge node was deleted through K8S
//It needs to be forcibly removed from the service topology and swirl
func deleteEdgeNode(nodekey interface{}) {
	node := nodekey.(*corev1.Node)

	nodeID := "" //node.Name
	//node.Name
	for _, addr := range node.Status.Addresses {
		//if addr.Type == v1.NodeExternalIP {
		//	nodeID = addr.Address
		if addr.Type == v1.NodeInternalIP && nodeID == "" {
			nodeID = addr.Address
		}
	}

	algorithm.Clusterer.RemoveEdgeNode(config.Cfg.MaxPing, nodeID, true)
}

func deleteDeployment(podkey interface{}) {
	pod := podkey.(*corev1.Pod)
	nodeID := pod.Spec.NodeName //.ObjectMeta.OwnerReferences[0].Name
	algorithm.Clusterer.RemoveDeployment(config.Cfg.MaxPing, nodeID, true)
}

func deleteFogNode(nodekey interface{}) {
	//eh.. hmm. Not sure how to do this yet, it involves kicking a lot of edge nodes to other fog nodes. Might not even be possible
	node := nodekey.(*corev1.Node)
	nodeID := ""
	for _, addr := range node.Status.Addresses {
		//if addr.Type == v1.NodeExternalIP {
		//	nodeID = addr.Address
		if addr.Type == v1.NodeInternalIP && nodeID == "" {
			nodeID = addr.Address
		}
	}

	algorithm.Clusterer.RemoveFogNode(nodeID, config.Cfg.MaxPing)
	//and this
	ws.UpdateFogNodeLists()
}

func initFogDeployment(nodeID string) {
	node := getK8SNode(nodeID)

	deployment, err := getDefaultDeployment()
	if err != nil {
		return
	}

	labels := make(map[string]string)
	labels["kubernetes.io/hostname"] = node.Labels["kubernetes.io/hostname"]
	deployment.Spec.Template.Spec.NodeSelector = labels
	deployment.Name = fmt.Sprintf("%s%s", deployment.Name, nodeID)

	k8sClient.AppsV1().Deployments("default").Create(deployment)
}

func deleteFogDeployment(nodeID string) {
	//node := getK8SNode(nodeID)

	deployment, err := getDefaultDeployment()
	if err != nil {
		return
	}

	depName := fmt.Sprintf("%s%s", deployment.Name, nodeID)

	k8sClient.AppsV1().Deployments("default").Delete(depName, &metav1.DeleteOptions{})
}

func getDefaultDeployment() (*appsv1.Deployment, error) {
	file, err := os.Open("deployment.json")
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	var deployment = &appsv1.Deployment{}
	err = decoder.Decode(deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func getK8SNode(ip string) *v1.Node {
	nodes, _ := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if nodes == nil || len(nodes.Items) == 0 {
		return nil
	}

	var foundNode *v1.Node
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Address == ip {
				foundNode = &node
			}
		}
	}
	return foundNode
}
