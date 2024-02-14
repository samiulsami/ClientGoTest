package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
)

var name string = "gobookstoreapi"

func CreateDeployment(client *dynamic.DynamicClient) {
	deploymentRes := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	deploymentObject := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"replicas": 2,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": name,
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": name,
						},
					},
					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  name,
								"image": "sami7786/gobookstoreapi:latest",
								"ports": []map[string]interface{}{
									{
										"name":          "http",
										"protocol":      "TCP",
										"containerPort": 3000,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	deployment, err := client.Resource(deploymentRes).Namespace("default").Create(context.TODO(), deploymentObject, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create deploymetn: %v\n", err)
	}
	fmt.Printf("Created deployment: %s\n", deployment.GetName())
}

func CreateService(client *dynamic.DynamicClient) {
	serviceRes := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	serviceObject := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": name + "-service",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": name,
				},
				"type": "LoadBalancer",
				"ports": []map[string]interface{}{
					{
						"protocol":   "TCP",
						"port":       3000,
						"targetPort": 3000,
						"nodePort":   30000,
					},
				},
			},
		},
	}

	service, err := client.Resource(serviceRes).Namespace("default").Create(context.TODO(), serviceObject, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create service: %v\n", err)
	}
	fmt.Printf("Service created: %v\n", service.GetName())
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	CreateDeployment(clientset)
	CreateService(clientset)
}
