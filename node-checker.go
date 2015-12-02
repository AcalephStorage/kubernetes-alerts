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
	*KVClient
	*NotifManager
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
	// process Node OOD
	// ...
}

func (n *NodeChecker) processNodeCheckReady(nodes []Node) {
	logrus.Info("Checking Node Readiness...")
	for _, node := range nodes {
		ready := false
		passThreshold := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == ConditionTypeReady {
				ready = condition.Status == "True"
				duration := time.Since(condition.LastTransitionTime)
				passThreshold = duration >= n.Threshold
			}
		}

		// node readiness may have changed
		if passThreshold {

			var message string
			var status CheckStatus
			if ready {
				status = CheckStatusPass
				message = node.Metadata.Name + " is Ready"
			} else {
				status = CheckStatusFail
				message = node.Metadata.Name + " is NOT Ready"
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

			exists, err := n.checkExists(check)
			if err != nil {
				logrus.WithError(err).Error("unable to determine if check exists or not")
				continue
			}
			if !exists {
				logrus.Infof("check %s is not in the record. recoding now", check.Name)
				err := n.saveCheck(check)
				if err != nil {
					logrus.WithError(err).Warnf("Unable to save check")
					continue
				}
				if check.Status == CheckStatusFail {
					logrus.Infof("check %s is new and failing, will notify", check.Name)
					n.addNotification(check)
				}
			} else {
				oldCheck, err := n.getCheck(check.CheckGroup, check.CheckType, check.Name)
				if err != nil {
					logrus.WithError(err).Warnf("unable to get previous check, can't proceed")
					continue
				}
				logrus.Printf("old: %s, new: %s", oldCheck.Status, check.Status)
				if check.Status != oldCheck.Status {
					logrus.Info("check %s status has changed, will notify", check.Name)
					logrus.Infof("status for %s:%s:%s has changed.", check.CheckGroup, check.CheckType, check.Name)
					err := n.saveCheck(check)
					if err != nil {
						logrus.WithError(err).Warnf("Unable to save")
						continue
					}
					n.addNotification(check)
				} else {
					logrus.Info("nothing has changed.")
				}
			}
		}

	}
}
