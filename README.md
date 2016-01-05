# kubernetes-alerts

Monitor Kubernetes and send alert notifications to Email, Slack etc

## Build

Clone the repository:

```
$ git clone git@github.com/AcalephStorage/kubernetes-alerts.git
$ cd kubernetes-alerts
```

Building the binary:

```
$ gb vendor restore
$ gb build
```

## TODO

Several TODOs at the moment:

- [ ] complete node checker
- [ ] create cluster checker
- [ ] create pod checker
- [ ] port notifiers from consul-alerts
- [ ] configuration via yaml
- ... more?
