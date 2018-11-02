// Copyright (c) 2018 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package calico

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type routeGenerator struct {
	client     *client
	kubeconfig string
}

func NewRouteGenerator(c *client, kubeconfig string) (rg *routeGenerator, err error) {
	rg = &routeGenerator{
		client:     c,
		kubeconfig: kubeconfig,
	}
	return
}

func (rg *routeGenerator) Start() (err error) {
	if rg.kubeconfig != "" {
		// Dynamic implementation based on currently active k8s Services and Endpoints.
		k8sClientset, err := getClients(rg.kubeconfig)
		if err != nil {
			log.WithError(err).Error("Failed to create k8s clientset")
			return err
		}
	} else {
		// MVP implementation: read CIDRs to advertise,
		// comma-separated, from an environment variable
		// CALICO_STATIC_ROUTES.
		routeString := os.Getenv("CALICO_STATIC_ROUTES")
		cidrs := []string{}
		for _, route := range strings.Split(routeString, ",") {
			cidr := strings.TrimSpace(route)
			if cidr != "" {
				cidrs = append(cidrs, cidr)
			}
		}
		rg.client.updateRoutes(cidrs)
	}
	return
}

// Stolen from kube-controllers: getClients builds and returns Kubernetes clients.
func getClients(kubeconfig string) (*kubernetes.Clientset, error) {

	// Build Kubernetes config; we support in-cluster config and kubeconfig as means of
	// configuring the client.
	k8sconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes client config: %s", err)
	}

	// Get Kubernetes clientset.
	k8sClientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes client: %s", err)
	}

	return k8sClientset, nil
}
