package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"
	"os"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/create"
)

func main() {
	configFile := "./kind.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	options := create.WithConfigFile(configFile)
	prgCluster := "kluster"
	ctx := cluster.NewContext(prgCluster)
	if err := ctx.Create(options); err != nil {
		log.Fatalf("err: %#v\n", err)
	}

	fmt.Printf("ctx.KubeConfigPath(): %s\n", ctx.KubeConfigPath())

	clientset, err := getClientSet(ctx.KubeConfigPath())
	if err != nil {
		log.Fatalf("err: %#v\n", err)
	}

	jobYAML, err := ioutil.ReadFile("job.yaml")
	if err != nil {
		log.Fatalf("err: %#v\n", err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(jobYAML), len(jobYAML))
	if err != nil {
		log.Fatalf("err: %#v\n", err)
	}

	job := &batchv1.Job{}
	if err := decoder.Decode(job); err != nil {
		log.Fatalf("err: %#v\n", err)
	}

	// Create Deployment
	fmt.Println("Creating Job...")
	_, err = clientset.BatchV1().Jobs(apiv1.NamespaceDefault).Create(job)
	if err != nil {
		log.Fatalf("err: %#v\n", err)
	}

	clientset, err = getClientSet(ctx.KubeConfigPath())
	if err != nil {
		log.Fatalf("err: %#v\n", err)
	}

	p, err := findPodForJob(clientset, "kube-bench")
	if err != nil {
		log.Fatalf("err: %#v\n", err)
	}

	output := getPodLogs(clientset, p)
	fmt.Printf("Output: %s\n", output)
}

func getClientSet(configPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func int32Ptr(i int32) *int32 { return &i }

func findPodForJob(clientset *kubernetes.Clientset, name string) (*apiv1.Pod, error) {

	for {
		pods, err := clientset.CoreV1().Pods(apiv1.NamespaceDefault).List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			fmt.Printf("Pod: %s\n", pod.Name)
			if strings.HasPrefix(pod.Name, name) {
				if  pod.Status.Phase == apiv1.PodSucceeded {
						return &pod, nil
				}
				time.Sleep(10 * time.Second)			
			}
		}
	}
	
	return nil, fmt.Errorf("no Pod with %s", name)
}

func getPodLogs(clientset *kubernetes.Clientset, pod *apiv1.Pod) string {
	podLogOpts := corev1.PodLogOptions{}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		return "error in opening stream"
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "error in copy information from podLogs to buf"
	}
	str := buf.String()

	return str
}
