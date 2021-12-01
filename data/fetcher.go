package data

import (
	"context"
	"fmt"
	"sync"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterState struct {
	Context string
	Host    string
	Error   error
	Nodes   *[]v1.Node
	Pods    *[]v1.Pod
}

func GetClusterState() ([]ClusterState, error) {
	cf := genericclioptions.ConfigFlags{}
	cc := cf.ToRawKubeConfigLoader()
	rawConfig, err := cc.RawConfig()
	if err != nil {
		return []ClusterState{}, fmt.Errorf("error loading kubeconfig: %s", err)
	}

	clusters := make([]ClusterState, 0, len(rawConfig.Contexts))
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for n := range rawConfig.Contexts {
		fmt.Printf("Fetching data from %s\n", n)
		wg.Add(1)
		go func(ctx string) {
			defer wg.Done()
			restConfig, clientSet, err := getKubeClient(ctx)
			if err != nil {
				reterr := fmt.Errorf("error getting client for %s: %s", ctx, err)
				clusters = append(clusters, ClusterState{Error: reterr, Context: ctx})
			}

			nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				reterr := fmt.Errorf("error getting nodes for %s: %s", ctx, err)
				clusters = append(clusters, ClusterState{Error: reterr, Context: ctx})
			}

			pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				reterr := fmt.Errorf("error getting nodes for %s: %s", ctx, err)
				clusters = append(clusters, ClusterState{Error: reterr, Context: ctx})
			}
			mu.Lock()
			defer mu.Unlock()
			clusters = append(clusters, ClusterState{Context: ctx, Host: restConfig.Host, Nodes: &nodes.Items, Pods: &pods.Items})
		}(n)
	}

	wg.Wait()
	return clusters, nil
}

func getKubeClient(context string) (*rest.Config, kubernetes.Interface, error) {
	config, err := configForContext(context)
	if err != nil {
		return nil, nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}

// configForContext creates a Kubernetes REST client configuration for a given kubeconfig context.
func configForContext(context string) (*rest.Config, error) {
	config, err := getConfig(context).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %s", context, err)
	}
	return config, nil
}

// getConfig returns a Kubernetes client config for a given context.
func getConfig(context string) clientcmd.ClientConfig {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if context != "" {
		overrides.CurrentContext = context
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
}
