package main

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	ConditionTypeReady     = "Ready"
	ConditionTypeOutOfDisk = "OutOfDisk"

	NodeCheckReady     = "NodeReady"
	NodeCheckOutOfDisk = "NodeOutOfDisk"
	NodeCheckCpu       = "NodeCpu"
	NodeCheckMem       = "NodeMem"
)

type NodeChecker struct {
	*KubernetesApi
	*HeapsterModelApi
	*EtcdApi
	RunWaitGroup  sync.WaitGroup
	CheckInterval time.Duration
	stopChannel   chan bool
	Threshold     time.Duration
}

func (n *NodeChecker) start() {
	logrus.Info("Starting Node Checker...")
	n.RunWaitGroup.Add(1)
	n.stopChannel = make(chan bool)
	go n.run()
}

func (n *NodeChecker) stop() {
	close(n.stopChannel)
	n.RunWaitGroup.Done()
}

func (n *NodeChecker) run() {
	running := true
	for running {
		select {
		case <-time.After(n.CheckInterval):
			n.processNodeCheck()
		case <-n.stopChannel:
			running = false
		}
		time.Sleep(1 * time.Second)
	}
}

func (n *NodeChecker) processNodeCheck() {
	logrus.Info("Running Node Checks...")
	nodes, err := n.Nodes()
	if err != nil {
		logrus.WithError(err).Error("Unable to retrieve nodes.")
		return
	}
	n.processNodeCheckReady(nodes)
}

func (n *NodeChecker) processNodeCheckReady(nodes []Node) {
	for _, node := range nodes {
		ready := false
		passThreshold := false
		message := ""
		for _, condition := range node.Status.Conditions {
			if condition.Type == ConditionTypeReady {
				ready = condition.Status == "True"
				duration := time.Since(condition.LastTransitionTime)
				passThreshold = duration > n.Threshold
				message = condition.Message
			}
		}

		// node readiness may have changed
		if passThreshold {

			var status CheckStatus
			if ready {
				status = CheckStatusPass
			} else {
				status = CheckStatusFail
			}

			check := KubeCheck{
				Name:       node.Metadata.Name,
				CheckGroup: CheckGroupNode,
				CheckType:  CheckTypeNodeReady,
				Status:     status,
				Message:    message,
				Timestamp:  time.Now(),
				Labels:     node.Metadata.Labels,
			}

			exists := n.checkExists(check)
			if !exists {
				err := n.EtcdApi.saveCheck(check)
				if err != nil {
					logrus.WithError(err).Warnf("Unable to save")
					continue
				}
				if check.Status == CheckStatusFail {
					// send notification if failed status
				}
			} else {
				oldCheck, err := n.EtcdApi.getCheck(check.CheckGroup, check.CheckType, check.Name)
				if err != nil {
					logrus.WithError(err).Warnf("unable to get previous check, can't proceed")
					continue
				}
				if check.Status != oldCheck.Status {
					logrus.Infof("status for %s:%s:%s has changed.", check.CheckGroup, check.CheckType, check.Name)
					// send notification since status has changed
					err := n.EtcdApi.saveCheck(check)
					if err != nil {
						logrus.WithError(err).Warnf("Unable to save")
						continue
					}
					// send notif here
				} else {
					logrus.Info("nothing has changed.")
				}
			}
		}

	}
}
