package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type restService struct {
	client   *http.Client
	endpoint string
}

func newRestService(config Config) (*restService, error) {
	client, err := httpClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client: %v", err)
	}
	return &restService{client, extractEndpoint(config)}, nil
}

func extractEndpoint(config Config) string {
	endpoint := config.Connection.APIEndpoint
	if strings.HasSuffix(endpoint, "/") {
		return endpoint
	}
	return endpoint + "/"
}

func httpClient(config Config) (*http.Client, error) {
	// load client certificate
	cert, err := tls.LoadX509KeyPair(config.Connection.Cert, config.Connection.Key)
	if err != nil {
		return nil, fmt.Errorf("error loading X509 key pair: %v", err)
	}
	// load CA file to verify server
	caPool := x509.NewCertPool()
	severCert, err := ioutil.ReadFile(config.Connection.CACert)
	if err != nil {
		return nil, fmt.Errorf("could not load CA file: %v", err)
	}
	caPool.AppendCertsFromPEM(severCert)
	// create a client with specific transport configurations
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      caPool,
			Certificates: []tls.Certificate{cert},
		},
	}
	client := &http.Client{Transport: transport}

	return client, nil
}

func (r *restService) Post(path string, json []byte) ([]byte, int, error) {
	output := strings.NewReader(string(json))
	log.Printf("debug: post request path: %s , body: %s", r.endpoint+path, string(json))
	response, err := r.client.Post(r.endpoint+path, "application/json", output)
	if err != nil {
		return nil, -1, fmt.Errorf("error on http request (POST %v): %v", r.endpoint+path, err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("error reading http request body: %v", err)
	}

	return body, response.StatusCode, err
}

func (r *restService) Delete(path string) ([]byte, int, error) {
	request, err := http.NewRequest("DELETE", r.endpoint+path, nil)
	if err != nil {
		return nil, -1, fmt.Errorf("error creating http request (DELETE %v): %v", r.endpoint+path, err)
	}
	response, err := r.client.Do(request)
	if err != nil {
		return nil, -1, fmt.Errorf("error executing http request (DELETE %v): %v", r.endpoint+path, err)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("error reading http request body: %v", err)
	}

	return body, response.StatusCode, nil
}

func (r *restService) Get(path string) ([]byte, int, error) {
	response, err := r.client.Get(r.endpoint + path)
	if err != nil {
		return nil, -1, fmt.Errorf("error on http request (GET %v): %v", r.endpoint+path, err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("error reading http request body: %v", err)
	}

	return body, response.StatusCode, nil
}
