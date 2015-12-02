# kubernetes-alerts

Monitor Kubernetes and send alert notifications to Email, Slack etc

## Build

To build, make sure you `export GO15VENDOREXPERIMENT=1` to enable the Go's vendoring.

Note: There's a slight difference in working with Go Vendoring in Linux and OS X, in Linux, the source should exist in `$GOPATH/src/github.com/AcalephStorage/kubernetes-alerts` while in OS X, the project can exist anywhere.

Clone the repository:

```
$ git clone git@github.com/AcalephStorage/kubernetes-alerts.git
$ cd kubernetes-alerts
$ git submodule init
$ git submodule update
```

Building the binary:

```
$ go build
```

## TODO

Several TODOs at the moment:

- [ ] complete node checker
- [ ] create cluster checker
- [ ] create pod checker
- [ ] port notifiers from consul-alerts
- [ ] configuration via yaml
- ... more?
