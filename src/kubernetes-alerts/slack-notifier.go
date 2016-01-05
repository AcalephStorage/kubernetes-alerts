package main

import (
	"bytes"
	"fmt"
	"strings"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
)

type SlackNotifier struct {
	ClusterName string       `json:"-"`
	Url         string       `json:"-"`
	Channel     string       `json:"channel"`
	Username    string       `json:"username"`
	IconUrl     string       `json:"icon_url"`
	IconEmoji   string       `json:"icon_emoji"`
	Text        string       `json:"text,omitempty"`
	Attachments []attachment `json:"attachments,omitempty"`
	Detailed    bool         `json:"-"`
}

type attachment struct {
	Color    string   `json:"color"`
	Title    string   `json:"title"`
	Pretext  string   `json:"pretext"`
	Text     string   `json:"text"`
	MrkdwnIn []string `json:"mrkdwn_in"`
}

func (slack *SlackNotifier) Notify(checks []KubeCheck) bool {
	logrus.Infof("Sending %d notifications to slack", len(checks))

	if slack.Detailed {
		return slack.notifyDetailed(checks)
	} else {
		return slack.notifySimple(checks)
	}

}

func (slack *SlackNotifier) notifySimple(checks []KubeCheck) bool {

	_, pass, warn, fail := NotifSummary(checks)

	textTemplate := `%s Notifications
	--------------------------------------------------------------------------------
	pass: %d warn: %d fail: %d
	--------------------------------------------------------------------------------
	%s
	--------------------------------------------------------------------------------
	`

	detailTemplate := ` [%s] %s: %s.\n`
	var details string
	for _, check := range checks {
		details += fmt.Sprintf(detailTemplate, strings.ToUpper(string(check.Status)), check.Timestamp.String(), check.Message)
	}

	text := fmt.Sprintf(textTemplate, slack.ClusterName, pass, warn, fail, details)

	slack.Text = text
	return slack.postToSlack()
}

func (slack *SlackNotifier) notifyDetailed(checks []KubeCheck) bool {

	overall, pass, warn, fail := NotifSummary(checks)

	var emoji, color string
	switch overall {
	case CheckStatusPass:
		emoji = ":white_check_mark:"
		color = "good"
	case CheckStatusWarn:
		emoji = ":question:"
		color = "warning"
	case CheckStatusFail:
		emoji = ":x:"
		color = "danger"
	default:
		emoji = ":question:"
	}

	title := "Kubernetes Alerts"

	preTextTemplate := `%s %s Notifications
	--------------------------------------------------------------------------------
	 %d :simple_smile: %d :fearful: %d :rage:
	--------------------------------------------------------------------------------
	`

	preText := fmt.Sprintf(preTextTemplate, emoji, slack.ClusterName, pass, warn, fail)

	detailTemplate := " %s %s: %s.\n"
	var details string
	for _, check := range checks {
		var statusEmoji string
		switch check.Status {
		case CheckStatusPass:
			statusEmoji = ":simple_smile:"
		case CheckStatusWarn:
			statusEmoji = ":fearful:"
		case CheckStatusFail:
			statusEmoji = ":rage:"
		}
		details += fmt.Sprintf(detailTemplate, statusEmoji, check.Timestamp.String(), check.Message)
		details += "\n"
	}

	a := attachment{
		Color:    color,
		Title:    title,
		Pretext:  preText,
		Text:     details,
		MrkdwnIn: []string{"text", "pretext"},
	}
	slack.Attachments = []attachment{a}

	return slack.postToSlack()

}

func (slack *SlackNotifier) postToSlack() bool {

	data, err := json.Marshal(slack)
	logrus.Println(string(data))
	if err != nil {
		logrus.Println("Unable to marshal slack payload:", err)
		return false
	}
	logrus.Debugf("struct = %+v, json = %s", slack, string(data))

	b := bytes.NewBuffer(data)
	if res, err := http.Post(slack.Url, "application/json", b); err != nil {
		logrus.Println("Unable to send data to slack:", err)
		return false
	} else {
		defer res.Body.Close()
		statusCode := res.StatusCode
		if statusCode != 200 {
			body, _ := ioutil.ReadAll(res.Body)
			logrus.Println("Unable to notify slack:", string(body))
			return false
		} else {
			logrus.Println("Slack notification sent.")
			return true
		}
	}

}
