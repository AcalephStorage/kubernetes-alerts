package main

import (
	"fmt"
)

type EtcdApi struct {
	*ApiClient
}

func (e *EtcdApi) checkExists(check KubeCheck) bool {
	endpoint := fmt.Sprintf("%s/keys/kube-alerts/%s/%s/%s", e.apiBaseUrl, check.CheckGroup, check.CheckType, check.Name)
	var data map[string]interface{}
	err := e.GetRequest(endpoint, data)
	if err != nil {
		// log warning
		return false
	}
	_, missing := data["errorCode"]
	return !missing
}

func (e *EtcdApi) saveCheck(check KubeCheck) error {
	endpoint := fmt.Sprintf("%s/keys/kube-alerts/%s/%s/%s", e.apiBaseUrl, check.CheckGroup, check.CheckType, check.Name)
	data, err := toReader(check)
	if err != nil {
		return err
	}
	if !e.checkExists(check) {
		return e.PostRequest(endpoint, data)
	}
	return e.PutRequest(endpoint, data)
}

func (e *EtcdApi) getCheck(checkGroup KubeCheckGroup, checkType KubeCheckType, checkName string) (KubeCheck, error) {
	endpoint := fmt.Sprintf("%s/keys/kube-alerts/%s/%s/%s", e.apiBaseUrl, checkGroup, checkType, checkName)
	var check KubeCheck
	err := e.GetRequest(endpoint, &check)
	if err != nil {
		return KubeCheck{}, err
	}
	return check, nil
}
