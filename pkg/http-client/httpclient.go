/*
Copyright © 2021 VMware
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	defaultTimeoutSeconds = 10
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// creates a new instance of Client, with default configuration
func New() *http.Client {
	// Return Transport for testing purposes
	return &http.Client{
		Timeout: time.Second * defaultTimeoutSeconds,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

// creates a new instance of Client, given a path to addtional certificates
// certFile may be empty string, which means no additional certs will be used
func NewWithCertFile(certFile string, skipTLS bool) (*http.Client, error) {
	// If additionalCA exists, load it
	if _, err := os.Stat(certFile); !os.IsNotExist(err) {
		certs, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf("failed to append %s to RootCAs: %v", certFile, err)
		}
		return NewWithCertBytes(certs, skipTLS)
	}

	// Return Transport for testing purposes
	client := New()
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLS,
		}}
	return client, nil
}

// creates a new instance of Client, given bytes for addtional certificates
func NewWithCertBytes(certs []byte, skipTLS bool) (*http.Client, error) {
	// Get the SystemCertPool, continue with an empty pool on error
	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}

	// Append our cert to the system pool
	if ok := caCertPool.AppendCertsFromPEM(certs); !ok {
		return nil, fmt.Errorf("failed to append bytes to RootCAs")
	}

	// Return Transport for testing purposes
	client := New()
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLS,
			RootCAs:            caCertPool,
		}}
	return client, nil
}

// performs an HTTP GET request using provided client, URL and request headers.
// returns response body, as bytes on successful status, or error body,
// if applicable on error status
func Get(url string, cli Client, headers map[string]string) ([]byte, error) {
	reader, _, err := GetStream(url, cli, headers)
	if reader != nil {
		defer reader.Close()
	}
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

// performs an HTTP GET request using provided client, URL and request headers.
// returns response body, as bytes on successful status, or error body,
// if applicable on error status
// returns response as a stream, as well as response content type
// NOTE: it is the caller's responsibility to close the reader stream when no longer needed
func GetStream(url string, cli Client, reqHeaders map[string]string) (io.ReadCloser, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	for header, content := range reqHeaders {
		req.Header.Set(header, content)
	}

	res, err := cli.Do(req)
	if err != nil {
		return nil, "", err
	}

	respContentType := res.Header.Get("Content-Type")

	if res.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("GET request to [%s] failed due to status [%d]", url, res.StatusCode)
		errPayload, err := ioutil.ReadAll(res.Body)
		if err == nil && len(errPayload) > 0 {
			errorMsg += ": " + string(errPayload)
		}
		return nil, respContentType, fmt.Errorf(errorMsg)
	}

	return res.Body, respContentType, nil
}
