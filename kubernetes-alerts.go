package main

import (
	"flag"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	CheckGroupCluster = KubeCheckGroup("cluster")
	CheckGroupNode    = KubeCheckGroup("node")
	CheckGroupPod     = KubeCheckGroup("pod")

	CheckTypeNodeReady     = KubeCheckType("node-ready")
	CheckTypeNodeOutOfDisk = KubeCheckType("node-out-of-disk")
	CheckTypeNodeCpu       = KubeCheckType("node-cpu")
	CheckTypeNodeMem       = KubeCheckType("node-mem")

	CheckStatusPass = CheckStatus("pass")
	CheckStatusWarn = CheckStatus("warn")
	CheckStatusFail = CheckStatus("fail")
)

type KubeCheckGroup string
type KubeCheckType string
type CheckStatus string

type KubeCheck struct {
	Name       string            `json:"name"`
	CheckGroup KubeCheckGroup    `json:"checkGroup"`
	CheckType  KubeCheckType     `json:"checkType"`
	Status     CheckStatus       `json:"status"`
	Message    string            `json:"message"`
	Timestamp  time.Time         `json:"timestamp"`
	Labels     map[string]string `json:"labels"`
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	kubernetes := &KubernetesApi{ApiClient: &ApiClient{}}
	heapster := &HeapsterModelApi{ApiClient: &ApiClient{}}
	etcd := &EtcdApi{ApiClient: &ApiClient{}}
	parseFlags(kubernetes, heapster, etcd)

	if err := kubernetes.prepareClient(); err != nil {
		logrus.WithError(err).Error("unable to create kubernetes client")
		os.Exit(-1)
	}

	if err := heapster.prepareClient(); err != nil {
		logrus.WithError(err).Error("unable to create heapster client")
		os.Exit(-1)
	}

	if err := etcd.prepareClient(); err != nil {
		logrus.WithError(err).Error("unable to create etcd client")
		os.Exit(-1)
	}

	nodeChecker := &NodeChecker{
		KubernetesApi:    kubernetes,
		HeapsterModelApi: heapster,
		EtcdApi:          etcd,
		CheckInterval:    5 * time.Second,
		Threshold:        1 * time.Minute,
	}

	logrus.Info("Starting kube-alerts...")

	nodeChecker.start()

	nodeChecker.RunWaitGroup.Wait()
}

func parseFlags(kubernetes *KubernetesApi, heapster *HeapsterModelApi, etcd *EtcdApi) {
	flag.StringVar(&kubernetes.apiBaseUrl, "k8s-api", "", "Kubernetes API Base URL")
	flag.StringVar(&kubernetes.certificateAuthority, "k8s-certificate-authority", "", "Kubernetes Certificate Authority")
	flag.StringVar(&kubernetes.clientCertificate, "k8s-client-certificate", "", "Kubernetes Client Certificate")
	flag.StringVar(&kubernetes.clientKey, "k8s-client-key", "", "Kubernetes Client Key")
	flag.StringVar(&kubernetes.token, "k8s-token", "", "Kubernetes Token")
	flag.StringVar(&heapster.apiBaseUrl, "heapster-api", "", "Heapster API Base URL")
	flag.StringVar(&heapster.certificateAuthority, "heapster-certificate-authority", "", "Heapster Certificate Authority")
	flag.StringVar(&heapster.clientCertificate, "heapster-client-certificate", "", "Heapster Client Certificate")
	flag.StringVar(&heapster.clientKey, "heapster-client-key", "", "Heapster Client Key")
	flag.StringVar(&heapster.token, "heapster-token", "", "Heapster Token")
	flag.StringVar(&etcd.apiBaseUrl, "etcd-api", "", "Etcd API Base URL")
	flag.StringVar(&etcd.certificateAuthority, "etcd-certificate-authority", "", "Etcd Certificate Authority")
	flag.StringVar(&etcd.clientCertificate, "etcd-client-certificate", "", "Etcd Client Certificate")
	flag.StringVar(&etcd.clientKey, "etcd-client-key", "", "Etcd Client Key")
	flag.Parse()
}
