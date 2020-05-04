package kube

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubeinformers "k8s.io/client-go/informers"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os"
	"swirl/swirlservice/algorithm"
	"swirl/swirlservice/config"
	"time"
)

var k8sClient *kubernetes.Clientset

var edgeInformer corev1informers.NodeInformer
var deployInformer corev1informers.PodInformer
var fogInformer corev1informers.NodeInformer

func Init() {
	clientset, err := GetKubeClient()
	if err != nil {
		fmt.Println("Failed to create k8s clientset")
		return
	}
	k8sClient = clientset

	edgeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(k8sClient, 1*time.Minute, kubeinformers.WithNamespace(corev1.NamespaceAll), kubeinformers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.LabelSelector = labels.Set(map[string]string{"swirlnodetype": config.Cfg.EdgeNodeLabel}).AsSelector().String()
	}))
	deployInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(k8sClient, 1*time.Minute, kubeinformers.WithNamespace(corev1.NamespaceAll), kubeinformers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.LabelSelector = labels.Set(map[string]string{"app": config.Cfg.EdgeDeploymentName}).AsSelector().String()
	}))
	fogInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(k8sClient, 1*time.Minute, kubeinformers.WithNamespace(corev1.NamespaceAll), kubeinformers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.LabelSelector = labels.Set(map[string]string{"swirlnodetype": config.Cfg.FogNodeLabel}).AsSelector().String()
	}))

	rootContext, _ := context.WithCancel(context.Background())

	edgeInformer = edgeInformerFactory.Core().V1().Nodes()
	go edgeInformerFactory.Start(rootContext.Done())
	deployInformer = deployInformerFactory.Core().V1().Pods()
	go deployInformerFactory.Start(rootContext.Done())
	fogInformer = fogInformerFactory.Core().V1().Nodes()
	go fogInformerFactory.Start(rootContext.Done())

	edgeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addEdgeNode,
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: deleteEdgeNode,
	})
	deployInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addDeployment,
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: deleteDeployment,
	})
	fogInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addFogNode,
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: deleteFogNode,
	})

	algorithm.ClusterInitialized = initFogDeployment
	algorithm.ClusterUninitialized = deleteFogDeployment
}

func GetKubeClient() (*kubernetes.Clientset, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	fmt.Printf("GetKubeClient: creating config with host %s port %s\n", host, port)

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	cfgJson, _ := json.Marshal(config)
	fmt.Printf("GetKubeClient: Config created %s\n", string(cfgJson))
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	clientsetnil := clientset == nil
	fmt.Printf("GetKubeClient: Clientset nil %t\n", clientsetnil)

	if err != nil {
		return clientset, err
	}

	//fmt.Println("GetKubeClient: Clientset created")
	return clientset, nil
}
