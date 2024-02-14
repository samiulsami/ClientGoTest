package main

import (
	"context"
	"flag"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

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
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	myDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gobookstoreapi",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "gobookstoreapi",
				},
			},
			Replicas: int32Ptr(2),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "gobookstoreapi",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "my-app",
							Image:           "sami7786/gobookstoreapi:latest",
							ImagePullPolicy: "IfNotPresent",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 3000,
								},
							},
						},
					},
				},
			},
		},
	}

	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)

	//todo
	//check this https://stackoverflow.com/questions/53874921/kubernetes-client-go-creating-services-and-enpdoints
	myService := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gobookstoreapi-service",
			Namespace: "default",
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"app": "gobookstoreapi",
			},
			Type: apiv1.ServiceTypeLoadBalancer,
			Ports: []apiv1.ServicePort{{
				Name:       "TCP",
				Port:       3000,
				TargetPort: intstr.FromInt32(3000),
				NodePort:   30000,
			},
			},
		},
	}

	result, err := deploymentsClient.Create(context.TODO(), myDeployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	result2, err2 := serviceClient.Create(context.TODO(), myService, metav1.CreateOptions{})
	if err != nil {
		panic(err2)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	fmt.Printf("Created service %q.\n", result2.GetObjectMeta().GetName())
}

func int32Ptr(i int32) *int32 { return &i }
