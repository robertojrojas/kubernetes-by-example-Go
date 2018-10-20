package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var kubeconfig *string
var namespace *string

func init() {
	fmt.Println("second init()")
}

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	namespace = flag.String("namespace", "default", "namespace to use")
	flag.Parse()

	fmt.Println("first init()")
}

func main() {

	clientset, err := getClientSet()
	check(err)

	pods, err := clientset.CoreV1().Pods(*namespace).List(metav1.ListOptions{})
	check(err)

	fmt.Printf("%-39s%-10s%-10s%-10s%s - (%d)\n", "NAME", "READY", "STATUS", "RESTARTS", "AGE", len(pods.Items))
	for _, pod := range pods.Items {
		cRunning := 0
		atLeastOneRunning := "Waiting"
		restarts := int32(0)
		oldestContainerRunning := time.Now()
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				cRunning++
				atLeastOneRunning = "Running"
			}
			restarts = restarts + cs.RestartCount
			if cs.State.Running.StartedAt.Time.Before(oldestContainerRunning) {
				oldestContainerRunning = cs.State.Running.StartedAt.Time
			}

		}
		readStr := fmt.Sprintf("%d/%d", cRunning, len(pod.Status.ContainerStatuses))

		fmt.Printf("%-39s%-10s%-10s%-10d%v\n", pod.Name, readStr, atLeastOneRunning, restarts, oldestContainerRunning)
	}

}

func getClientSet() (*kubernetes.Clientset, error) {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
