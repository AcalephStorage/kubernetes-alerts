package main

import (
	"errors"
	"io"
	"time"

	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
)

type ApiClient struct {
	*http.Client
	apiBaseUrl           string
	certificateAuthority string
	clientCertificate    string
	clientKey            string
	token                string
}

func (a *ApiClient) prepareClient() error {
	var cacert *x509.CertPool
	if a.certificateAuthority != "" {
		capem, err := ioutil.ReadFile(a.certificateAuthority)
		if err != nil {
			return err
		}
		cacert = x509.NewCertPool()
		if !cacert.AppendCertsFromPEM(capem) {
			return errors.New("unable to load certificate authority")
		}
	}

	var cert tls.Certificate
	if a.clientCertificate != "" && a.clientKey != "" {
		c := a.clientCertificate
		k := a.clientKey
		var err error
		cert, err = tls.LoadX509KeyPair(c, k)
		if err != nil {
			return err
		}
	}

	if cacert != nil || &cert != nil {
		config := &tls.Config{
			RootCAs:      cacert,
			Certificates: []tls.Certificate{cert},
		}
		transport := &http.Transport{
			TLSClientConfig:     config,
			TLSHandshakeTimeout: 5 * time.Second,
		}
		client := &http.Client{Transport: transport}
		a.Client = client
	} else {
		a.Client = &http.Client{}
	}
	return nil
}

func (a *ApiClient) GetRequest(path string, resData interface{}) error {
	endpoint := a.apiBaseUrl + path
	logrus.Debug("requesting", endpoint)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	if a.token != "" {
		req.Header.Add("Authorization", "Bearer "+a.token)
	}
	res, err := a.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, resData)
	if err != nil {
		return err
	}
	logrus.Debug("Successful.")
	return nil
}

func (a *ApiClient) PostRequest(path string, data io.Reader) error {
	endpoint := a.apiBaseUrl + path
	logrus.Debug("posting", endpoint)
	req, err := http.NewRequest("POST", endpoint, data)
	if err != nil {
		return err
	}
	if a.token != "" {
		req.Header.Add("Authorization", "Bearer "+a.token)
	}
	res, err := a.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New(res.Status)
}

func (a *ApiClient) PutRequest(path, data string) error {
	endpoint := a.apiBaseUrl + path
	logrus.Info("putting: ", endpoint)
	req, err := http.NewRequest("PUT", endpoint, nil)
	req.PostForm["value"] = []string{data}

	if err != nil {
		return err
	}
	if a.token != "" {
		req.Header.Add("Authorization", "Bearer "+a.token)
	}
	res, err := a.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	logrus.Println(res.StatusCode)
	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New(res.Status)
}
