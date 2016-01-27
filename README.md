kube-alerts
===========

[![Build Status](https://travis-ci.org/AcalephStorage/kubernetes-alerts.svg?branch=develop)](https://travis-ci.org/AcalephStorage/kubernetes-alerts)

Monitor Kubernetes and send alert notifications to Email, Slack etc.

This follows a similar approach to [consul-alerts](https://github.com/AcalephStorage/consul-alerts). 

## Requirement

1. Kubernetes
2. Heapster
3. Etcd

Releases
--------

Binaries are [here](https://github.com/AcalephStorage/kubernetes-alerts/releases) and docker images [here](https://quay.io/repository/acaleph/kube-alerts).

Build
-----

To build from source, clone the repo:

```
$ git clone https://github.com/AcalephStorage/kubernetes-alerts.git
$ cd kubernetes-alerts
```

Get gb (package manager):

```
$ go get github.com/constabulary/gb/...
```

Get dependencies:

```
$ make deps
```

and lastly, build:

```
$ make build
```

The binary will be in `bin/$GOOS/$GOARCH` directory.

Docker
------

The docker image can be pulled from quay.io.

```
$ docker pull quay.io/acaleph/kube-alerts:$tag
```

Usage
-----

```
$ kube-alerts [options]
```

or using docker:

```
$ docker run quay.io/acaleph/kube-alerts:$tag [options]
```

Configuration
-------------

### Connection

kube-alerts requires to connect to Kubernetes, Heapster, and ETCD. Here are the flags to configure the connections:

#### Kubernetes flags

| flag                       | description                                                                              | example                           |
|----------------------------|------------------------------------------------------------------------------------------|-----------------------------------|
| -k8s-api                   | the base url for the Kubernetes API                                                      | https://localhost/api/v1          |
| -k8s-certificate-authority | the certificate authority of the Kubernetes API                                          | /etc/kubernetes/ssl/ca.pem        |
| -k8s-client-certificate    | the client certificate for authentication                                                | /etc/kubernetes/ssl/admin.pem     |
| -k8s-client-key            | the client key for authentication                                                        | /etc/kubernetes/ssl/admin-key.pem |
| -k8s-token                 | the token for authentication                                                             | F0XBLTDaL3xDlBsq5YKAFIH7yzZNBhs6  |
| -k8s-token-file            | the file where token is stored. This flag is only considered if `-k8s-token` is not used | /path/to/token                    |

#### Heapster flags

| flag                            | description                                     | example                           |
|---------------------------------|-------------------------------------------------|-----------------------------------|
| -heapster-api                   | the base url for the Heapster API               | https://localhost/api/v1          |
| -heapster-certificate-authority | the certificate authority of the Heapster API   | /etc/kubernetes/ssl/ca.pem        |
| -heapster-client-certificate    | the client certificate for authentication       | /etc/kubernetes/ssl/admin.pem     |
| -heapster-client-key            | the client key for authentication               | /etc/kubernetes/ssl/admin-key.pem |
| -heapster-token                 | the token for authentication                    | F0XBLTDaL3xDlBsq5YKAFIH7yzZNBhs6  |

Note: Heapster can be accessed via Kubernetes. The heapster flag may change (not yet used).

#### KV store flags

| flag                      | description                                        | example                      |
|---------------------------|----------------------------------------------------|------------------------------|
| -kv-addresses             | comma separated addresses for the KV store         | localhost:2379               |
| -kv-backend               | the KV store backend (only etcd supported for now) | etcd                         |
| -kv-certificate-authority | the certificate authority of the KV store          | /etc/etcd/ssl/ca.pem         |
| -kv-client-certificate    | the client certificate for authentication          | /etc/etcd/ssl/client.pem     |
| -kv-client-key            | the client key for authentication                  | /etc/etcd/ssl/client-key.pem |


### Monitoring

There are three major kinds of checks that are monitored by kube-alerts. Node checks, cluster checks, and resource checks (pods). At the moment, only node checks are available. Here the options:

#### Node check flags

| flag                  | description                                                                  | example |
|-----------------------|------------------------------------------------------------------------------|---------|
| -node-check-interval  | interval when running the node checks (seconds)                              | 10      |
| -node-check-threshold | amount of time (seconds) a change of state needed to qualify as state change | 60      |


### Notification

Different notifiers can be configured. At the moment, only Slack and Email are supported.

#### General notification flags

| flag                   | description                                                           | example |
|------------------------|-----------------------------------------------------------------------|---------|
| -notification-interval | amount of time (seconds) to wait before sending pending notifications | 60      |
| -enable-email          | enable email notifier                                                 | true    |
| -enable-slack          | enable slack notifier                                                 | true    |

#### Email notifier flags

| flag                   | description                                             | example       |
|------------------------|---------------------------------------------------------|---------------|
| -email-cluster-name    | the cluster name to appear on the default email message | acaleph       |
| -email-url             | the SMTP server URL                                     | localhost     |     
| -email-port            | the SMTP server port                                    | 25            |
| -email-username        | the SMTP username                                       | user          |
| -email-password        | the SMTP password                                       | password      |
| -email-receivers       | comma-separated list of email to receive notifications  | dood@acale.ph |
| -email-sender-email    | the email of the sender                                 | food@acale.ph |
| -email-sender-alias    | alias of the sender                                     | kube-alerts   |
| -email-template        | custom email template                                   |               |

Note: The custom email template is optional, a default template will be used if this is not provided. (TODO: document custom template)

#### Slack notifier flags

| flag                | description                                             | example                              |
|---------------------|---------------------------------------------------------|--------------------------------------|
| -slack-cluster-name | the cluster name to appear on the default slack message | acaleph                              |
| -slack-url          | the slack webhook URL                                   | https://hooks.slack.com/services/... |
| -slack-username     | the username to appear on the slack message             | 25                                   |

### Logging

Log level can be set to limit the verbosity of the log.

| flag       | description                                                   | example |
|------------|---------------------------------------------------------------|---------|
| -log-level | log level, valid values are [debug, info, warn, error, panic] | debug   |

TODO
----

This is an initial release, a few more things needs to be done:

 - [ ] implement cluster level checks
 - [ ] implement pod/resource level checks
 - [ ] document email template
 - [ ] add more notifiers
 - [ ] simpler configuration (via YAML?)
 - [ ] Real tests

Contribution
------------

PRs are more than welcome. Just fork, create a feature branch, and open a PR. :)
