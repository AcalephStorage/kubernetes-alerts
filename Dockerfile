FROM golang:1.5
MAINTAINER Acaleph <admin@acale.ph>

ADD . /kube-alerts
RUN go get github.com/constabulary/gb/...

WORKDIR /kube-alerts
RUN gb vendor restore
RUN gb build

EXPOSE 9000
CMD []
ENTRYPOINT [ "/kube-alerts/bin/kubernetes-alerts" ]
