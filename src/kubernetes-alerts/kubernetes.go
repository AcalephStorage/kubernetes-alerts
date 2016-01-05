package main

import "time"

type KubernetesApi struct {
	*ApiClient
}

type NodeList struct {
	Items []Node `json:"items"`
}

type Node struct {
	Metadata ResourceMetadata `json:"metadata"`
	Status   NodeStatus       `json:"status"`
}

type ResourceMetadata struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type NodeStatus struct {
	Capacity   NodeCapacity    `json:"capacity"`
	Conditions []NodeCondition `json:"conditions"`
}

type NodeCapacity struct {
	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`
	Pods   string `json:"pods"`
}

type NodeCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastHeartbeatTime  time.Time `json:"lastHearbeatTime"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

func (k *KubernetesApi) Nodes() ([]Node, error) {
	var nodeList NodeList
	err := k.GetRequest("/nodes", &nodeList)
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}
