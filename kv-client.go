package main

import (
	"errors"
	"fmt"
	"time"

	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
)

type KVClient struct {
	backend              store.Backend
	addresses            []string
	certificateAuthority string
	clientCertificate    string
	clientKey            string
	store                store.Store
}

func (kvc *KVClient) prepareClient() error {
	hasCA := kvc.certificateAuthority != ""
	hasCert := kvc.clientCertificate != ""
	hasKey := kvc.clientKey != ""

	config := &store.Config{
		ConnectionTimeout: 5 * time.Second,
	}
	if hasCA || hasCert || hasKey {

		var cacert *x509.CertPool
		if kvc.certificateAuthority != "" {
			capem, err := ioutil.ReadFile(kvc.certificateAuthority)
			if err != nil {
				return err
			}
			cacert = x509.NewCertPool()
			if !cacert.AppendCertsFromPEM(capem) {
				return errors.New("unable to load certificate authority")
			}
		}

		var cert tls.Certificate
		if kvc.clientCertificate != "" && kvc.clientKey != "" {
			c := kvc.clientCertificate
			k := kvc.clientKey
			var err error
			cert, err = tls.LoadX509KeyPair(c, k)
			if err != nil {
				return err
			}
		}

		config.ClientTLS = &store.ClientTLSConfig{
			CertFile:   kvc.clientCertificate,
			KeyFile:    kvc.clientKey,
			CACertFile: kvc.certificateAuthority,
		}
		config.TLS = &tls.Config{
			RootCAs:      cacert,
			Certificates: []tls.Certificate{cert},
		}

	}
	store, err := libkv.NewStore(kvc.backend, kvc.addresses, config)
	if err != nil {
		fmt.Println(err)
		logrus.Error("unable to create kvclient. ", err)
		return err
	}
	kvc.store = store
	return nil
}

func (kvc *KVClient) checkExists(check KubeCheck) (bool, error) {
	key := fmt.Sprintf("kube-alerts/%s/%s/%s", check.CheckGroup, check.CheckType, check.Name)
	exists, err := kvc.store.Exists(key)
	if err != nil {
		logrus.WithError(err).Error("unable to check key existence")
		return false, err
	}
	return exists, nil
}

func (kvc *KVClient) saveCheck(check KubeCheck) error {
	value, err := json.Marshal(&check)
	if err != nil {
		logrus.WithError(err).Error("unable to marshall check")
		return err
	}
	key := fmt.Sprintf("kube-alerts/%s/%s/%s", check.CheckGroup, check.CheckType, check.Name)
	return kvc.store.Put(key, value, nil)
}

func (kvc *KVClient) getCheck(checkGroup KubeCheckGroup, checkType KubeCheckType, checkName string) (KubeCheck, error) {
	var check KubeCheck
	key := fmt.Sprintf("kube-alerts/%s/%s/%s", checkGroup, checkType, checkName)
	kvpair, err := kvc.store.Get(key)
	if err != nil {
		logrus.WithError(err).Error("unable to get kv pair")
		return check, err
	}
	err = json.Unmarshal(kvpair.Value, &check)
	if err != nil {
		logrus.WithError(err).Error("unable to unmarshal kv value")
		return check, err
	}
	return check, nil
}
