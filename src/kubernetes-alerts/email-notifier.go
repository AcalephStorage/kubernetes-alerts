package main

import (
	"bytes"
	"fmt"

	"html/template"
	"net/smtp"

	"github.com/Sirupsen/logrus"
)

type EmailNotifier struct {
	Enabled     bool
	ClusterName string
	Template    string
	Url         string
	Port        int
	Username    string
	Password    string
	SenderAlias string
	SenderEmail string
	Receivers   []string
}

type EmailData struct {
	ClusterName  string
	SystemStatus string
	FailCount    int
	WarnCount    int
	PassCount    int
	Nodes        map[string][]KubeCheck
}

func (email *EmailNotifier) Notify(checks []KubeCheck) bool {
	logrus.Infof("Sending %d notification email", len(checks))

	overall, pass, warn, fail := NotifSummary(checks)
	nodeMap := mapByNodes(checks)

	e := EmailData{
		ClusterName:  email.ClusterName,
		SystemStatus: string(overall),
		FailCount:    fail,
		WarnCount:    warn,
		PassCount:    pass,
		Nodes:        nodeMap,
	}

	var tmpl *template.Template
	var err error
	if email.Template == "" {
		tmpl, err = template.New("base").Parse(defaultTemplate)
	} else {
		tmpl, err = template.ParseFiles(email.Template)
	}

	if err != nil {
		logrus.WithError(err).Error("Invalid Template")
		return false
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, e); err != nil {
		logrus.WithError(err).Error("Unable to execute template")
		return false
	}

	msg := ""
	msg += fmt.Sprintf("From: \"%s\" <%s>\n", email.SenderAlias, email.SenderEmail)
	msg += fmt.Sprintf("Subject: %s is %s\n", email.ClusterName, overall)
	msg += "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg += body.String()

	addr := fmt.Sprintf("%s:%d", email.Url, email.Port)
	auth := smtp.PlainAuth("", email.Username, email.Password, email.Url)
	if err := smtp.SendMail(addr, auth, email.SenderEmail, email.Receivers, []byte(msg)); err != nil {
		logrus.WithError(err).Error("Unable to send notification.")
		return false
	}
	logrus.Infof("Email notification sent.")
	return true

}

func mapByNodes(checks []KubeCheck) map[string][]KubeCheck {
	nodeMap := make(map[string][]KubeCheck)
	for _, check := range checks {
		nodeName := check.Node
		nodeChecks := nodeMap[nodeName]
		if nodeChecks == nil {
			nodeChecks = make([]KubeCheck, 0)
		}
		nodeChecks = append(nodeChecks, check)
		nodeMap[nodeName] = nodeChecks
	}
	return nodeMap
}

var defaultTemplate string = `
<!DOCTYPE html>
<html lang="en">
	<head>
  		<title>{{ .ClusterName }}</title>
	</head>

	<body style="width:100% !important; min-width: 100%; -webkit-text-size-adjust:100%; -ms-text-size-adjust:100%; margin:0; padding:0; font-family: 'Helvetica', 'Arial', sans-serif; color: #000000;">

		<div style="margin-left: auto; margin-right: auto; width: 36em; padding: 10dp; font-weight: bold; color: #ffffff; background-color: {{ if .IsCritical }}#e13329{{ else if .IsWarning }}#eebb00{{ else if .IsPassing }}#24c75a{{ end }};">
			<div style="padding: 10px;">
				{{ .ClusterName }}
			</div>
		</div>

		<div style="margin-left: auto; margin-right: auto; width: 36em; margin-top: 10px; margin-bottom: 10px; padding: 10dp">
			<p>
			<span style="font-weight: bold; font-size: 1.05em;">System is {{ .SystemStatus }}</span>
			<br/>
			<span style="font-size: 0.9em;">The following nodes are currently experiencing issues:</span>
			<div style="font-size: 0.85em;">
				<div style="float: left; width: 33%;">
					<strong>Failed: </strong>
					<span>{{ .FailCount }}</span>
				</div>
				<div style="float: right; width: 33%;">
					<strong>Warning: </strong>
					<span>{{ .WarnCount }}</span>
				</div>
				<div style="display: inline-block; width: 33%;">
					<strong>Passed: </strong>
					<span>{{ .PassCount }}</span>
				</div>
			</div>
			</p>

		</div>

		{{ range $name, $checks := .Nodes }}
		<div style="margin-left: auto; margin-right: auto; width: 36em; padding-top: 5px; padding-bottom: 20px;">
			<div style="font-size: 1.1em;">
				<strong>Node: </strong>
				<strong>{{ $name }}</strong>
			</div>

			{{ range $check := $checks }}
			<div style="margin-top: 15px; padding: 10px; background-color: {{ if $check.IsCritical }}#e13329{{ else if $check.IsWarning }}#eebb00{{ else if $check.IsPassing }}#24c75a{{ end }};">
				<div style="font-weight: bold; font-size: 1.1em;">
					{{ with $check.Service }}
					{{ $check.Service }}:
					{{ end }}
					{{ $check.Check }}
				</div>
				<div style="font-size: 0.85em;">
					<strong>Since: </strong>
					<span>{{ $check.Timestamp }}</span>
				</div>
				{{ with $check.Notes }}
				<div style="padding-top: 15px;">
					<strong>Notes: </strong>
					<pre>{{ $check.Notes }}</pre>
				</div>
				{{end }}
				<div style="padding-top: 15px;">
					<strong>Output:</strong>
					<pre>{{ $check.Output }}</pre>
				</div>
			</div>
			{{ end }}

		</div>
		{{ end }}


	</body>

</html>
`
