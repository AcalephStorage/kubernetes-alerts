package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
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
	Node       string            `json:"node"`
	CheckGroup KubeCheckGroup    `json:"checkGroup,string"`
	CheckType  KubeCheckType     `json:"checkType,string"`
	Status     CheckStatus       `json:"status"`
	Message    string            `json:"message"`
	Timestamp  time.Time         `json:"timestamp"`
	Labels     map[string]string `json:"labels"`
}

func main() {
	// logrus.SetLevel(logrus.DebugLevel)

	kubernetes := &KubernetesApi{ApiClient: &ApiClient{}}
	heapster := &HeapsterModelApi{ApiClient: &ApiClient{}}
	kv := &KVClient{}
	slack := &SlackNotifier{Detailed: true}
	email := &EmailNotifier{}

	notifManager := &NotifManager{
		Notifiers: []Notifier{slack, email},
	}

	nodeChecker := &NodeChecker{
		KubernetesApi:    kubernetes,
		HeapsterModelApi: heapster,
		KVClient:         kv,
		NotifManager:     notifManager,
	}

	// need better way for configuring this...
	parseFlags(kubernetes, heapster, kv, notifManager, nodeChecker, slack, email)
	initLibKV()

	if err := kubernetes.prepareClient(); err != nil {
		logrus.WithError(err).Error("unable to create kubernetes client")
		os.Exit(-1)
	}

	if err := heapster.prepareClient(); err != nil {
		logrus.WithError(err).Error("unable to create heapster client")
		os.Exit(-1)
	}

	if err := kv.prepareClient(); err != nil {
		logrus.WithError(err).Error("unable to create kv client")
		os.Exit(-1)
	}

	logrus.Info("Starting kube-alerts...")

	notifManager.Start()
	nodeChecker.start()

	nodeChecker.RunWaitGroup.Wait()

	// clean up aka stop all services
}

func parseFlags(kubernetes *KubernetesApi, heapster *HeapsterModelApi, kv *KVClient, notifManager *NotifManager, nodeChecker *NodeChecker, slack *SlackNotifier, email *EmailNotifier) {
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

	flag.StringVar(&kv.certificateAuthority, "kv-certificate-authority", "", "KV Certificate Authority")
	flag.StringVar(&kv.clientCertificate, "kv-client-certificate", "", "KV Client Certificate")
	flag.StringVar(&kv.clientKey, "kv-client-key", "", "KV Client Key")

	notifIntervalSecs := flag.Int("notification-interval", 60, "the interval to wait before sending notifications (seconds)")

	nodeCheckIntervalSecs := flag.Int("node-check-interval", 10, "interval in seconds before running node checks")
	nodeCheckThresholdSecs := flag.Int("node-check-threshold", 60, "threshold before marking a node status as changed")

	flag.BoolVar(&slack.Enabled, "enable-slack", false, "Enable slack notifier")
	flag.StringVar(&slack.ClusterName, "slack-cluster-name", "", "Cluster name to display on slack notifications")
	flag.StringVar(&slack.Url, "slack-url", "", "The slack URL for notification")
	flag.StringVar(&slack.Username, "slack-username", "kube-alerts", "The slack username")

	flag.BoolVar(&email.Enabled, "enable-email", false, "Enable email notifier")
	flag.StringVar(&email.ClusterName, "email-cluster-name", "kubernetes", "The name of the kubernetes cluster")
	flag.StringVar(&email.Template, "email-template", "", "The email template file")
	flag.StringVar(&email.Url, "email-url", "", "The smtp server URL")
	flag.IntVar(&email.Port, "email-port", 0, "The smtp port")
	flag.StringVar(&email.Username, "email-username", "", "The smtp username")
	flag.StringVar(&email.Password, "email-password", "", "The smtp password")
	flag.StringVar(&email.SenderAlias, "email-sender-alias", "kube-alerts", "The email sender alias")
	flag.StringVar(&email.SenderEmail, "email-sender-email", "", "The email of the sender")

	emailReceivers := flag.String("email-receivers", "", "Comma separated list of receiver's email")
	email.Receivers = strings.Split(*emailReceivers, ",")

	addresses := flag.String("kv-addresses", "", "addresses for the KV store")
	backend := flag.String("kv-backend", "", "KV Store Backend. Can be etcd, consul, zk, boltdb")

	flag.Parse()

	kv.addresses = strings.Split(*addresses, ",")
	switch *backend {
	case "etcd":
		kv.backend = store.ETCD
	case "consul":
		kv.backend = store.CONSUL
	case "zk":
		kv.backend = store.ZK
	case "boltdb":
		kv.backend = store.BOLTDB
	}

	notifManager.NotifInterval = time.Duration(*notifIntervalSecs) * time.Second
	nodeChecker.CheckInterval = time.Duration(*nodeCheckIntervalSecs) * time.Second
	nodeChecker.Threshold = time.Duration(*nodeCheckThresholdSecs) * time.Second
}

func initLibKV() {
	etcd.Register()
}
