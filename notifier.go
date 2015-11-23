package main

import (
	"time"
)

type Notifier interface {
	Notify(check KubeCheck) error
}

type NotifManager struct {
	NotifChannel  chan KubeCheck
	notifInterval time.Duration
	Notifiers     []Notifier
	stopChannel   chan bool
	Checks        []KubeCheck
}

func (n *NotifManager) Start() {
	n.stopChannel = make(chan bool)
	// n.Checks = make([]KubeCheck, )
}

func (n *NotifManager) Stop() {
	running := true
	for running {
		select {
		case <-n.stopChannel:
			running = false
			break
		case <-time.After(n.notifInterval):
			n.sendNotifications()
		case <-n.NotifChannel:

		}
	}
}

func (n *NotifManager) sendNotifications() {

}
