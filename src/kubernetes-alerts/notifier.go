package main

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

type Notifier interface {
	Notify(checks []KubeCheck) bool
	NotifEnabled() bool
}

type NotifManager struct {
	NotifInterval      time.Duration
	Notifiers          []Notifier
	notifChannel       chan KubeCheck
	stopChannel        chan bool
	checks             []KubeCheck
	addCheckWaitGroup  sync.WaitGroup
	sendNotifWaitGroup sync.WaitGroup
}

func (n *NotifManager) Start() {
	logrus.Info("Starting notif manager...")
	n.notifChannel = make(chan KubeCheck, 10)
	n.stopChannel = make(chan bool)
	n.checks = make([]KubeCheck, 0)
	go n.listenForNotif()
}

func (n *NotifManager) Stop() {
	n.stopChannel <- true
	close(n.stopChannel)
	close(n.notifChannel)
}

func (n *NotifManager) listenForNotif() {
	running := true
	for running {
		select {
		case <-n.stopChannel:
			running = false
		case <-time.After(n.NotifInterval):
			logrus.Debug("Trying to send notifications...")
			n.addCheckWaitGroup.Wait()
			n.sendNotifWaitGroup.Add(1)
			n.sendNotifications()
			n.sendNotifWaitGroup.Done()
		case check := <-n.notifChannel:
			logrus.Debug("Adding check for notification...")
			n.sendNotifWaitGroup.Wait()
			n.addCheckWaitGroup.Add(1)
			n.checks = append(n.checks, check)
			n.addCheckWaitGroup.Done()
		}
	}
}

func (n *NotifManager) addNotification(check KubeCheck) {
	n.notifChannel <- check
}

func (n *NotifManager) sendNotifications() {
	if len(n.checks) > 0 {
		for _, notifier := range n.Notifiers {
			if notifier.NotifEnabled() {
				notifier.Notify(n.checks)
			}
		}
		n.checks = make([]KubeCheck, 0)
	}
}

func NotifSummary(checks []KubeCheck) (overall CheckStatus, pass, warn, fail int) {
	overall = CheckStatusPass
	for _, check := range checks {
		switch check.Status {
		case CheckStatusPass:
			pass++
		case CheckStatusWarn:
			warn++
			if overall != CheckStatusFail {
				overall = CheckStatusWarn
			}
		case CheckStatusFail:
			fail++
			overall = CheckStatusFail
		}
	}
	return
}
