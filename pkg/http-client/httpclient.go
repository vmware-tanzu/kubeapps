// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"golang.org/x/net/http/httpproxy"
	"io"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	defaultTimeoutSeconds = 180
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientWithDefaults implements Client interface
// and includes an override of the Do method which injects the following supported defaults:
//   - headers: e.g. User-Agent and Authorization (when present)
type ClientWithDefaults struct {
	Client         Client
	DefaultHeaders http.Header
}

// Do (in ClientWithDefaults) HTTP request
func (c *ClientWithDefaults) Do(req *http.Request) (*http.Response, error) {
	for k, v := range c.DefaultHeaders {
		// Only add the default header if it's not already set in the request.
		if _, ok := req.Header[k]; !ok {
			req.Header[k] = v
		}
	}
	return c.Client.Do(req)
}

// New creates a new instance of http Client, with following default configuration:
//   - timeout
//   - proxy from environment
func New() *http.Client {
	return &http.Client{
		Timeout: time.Second * defaultTimeoutSeconds,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

// NewWithCertFile creates a new instance of Client, given a path to additional certificates
// certFile may be empty string, which means no additional certs will be used
func NewWithCertFile(certFile string, skipTLS bool) (*http.Client, error) {
	// If additionalCA exists, load it
	if _, err := os.Stat(certFile); !os.IsNotExist(err) {
		// #nosec G304
		certs, err := os.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf("failed to append %s to RootCAs: %v", certFile, err)
		}
		return NewWithCertBytes(certs, skipTLS)
	}

	// Return client with TLS skipVerify but no additional certs
	client := New()
	// #nosec G402
	config := &tls.Config{
		InsecureSkipVerify: skipTLS,
	}
	if err := SetClientTLS(client, config); err != nil {
		return nil, err
	}

	return client, nil
}

// NewWithCertBytes creates a new instance of Client, given bytes for additional certificates
func NewWithCertBytes(certs []byte, skipTLS bool) (*http.Client, error) {
	// create cert pool
	caCertPool, err := GetCertPool(certs)
	if err != nil {
		return nil, err
	}

	// create and configure client
	client := New()
	// #nosec G402
	config := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: skipTLS,
	}
	if err := SetClientTLS(client, config); err != nil {
		return nil, err
	}

	return client, nil
}

// GetCertPool get or create a cert pool, with the given (optional) certs
func GetCertPool(certs []byte) (*x509.CertPool, error) {
	// Require the SystemCertPool unless the env var is explicitly set.
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		caCertPool = x509.NewCertPool()
	}

	// Append our cert to the system pool
	if len(certs) > 0 {
		if ok := caCertPool.AppendCertsFromPEM(certs); !ok {
			return nil, fmt.Errorf("failed to append certs to RootCAs")
		}
	}

	return caCertPool, nil
}

// SetClientProxy configure the given proxy on the given client
func SetClientProxy(client *http.Client, proxy func(*http.Request) (*url.URL, error)) error {
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		return fmt.Errorf("transport was not an http.Transport")
	}
	transport.Proxy = proxy
	return nil
}

// SetClientTLS configure the given tls on the given client
func SetClientTLS(client *http.Client, config *tls.Config) error {
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		return fmt.Errorf("transport was not an http.Transport")
	}
	if config == nil {
		return fmt.Errorf("invalid argument: config may not be nil")
	}
	transport.TLSClientConfig = config.Clone()
	return nil
}

func NewClientTLS(certBytes, keyBytes, caBytes []byte) (*tls.Config, error) {
	config := tls.Config{MinVersion: tls.VersionTLS12}

	if len(certBytes) != 0 && len(keyBytes) != 0 {
		cert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, err
		}
		config.Certificates = []tls.Certificate{cert}
	}

	if len(caBytes) != 0 {
		cp, err := GetCertPool(caBytes)
		if err != nil {
			return nil, err
		}
		config.RootCAs = cp
	}
	return &config, nil
}

// Get performs an HTTP GET request using provided client, URL and request headers.
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
	return io.ReadAll(reader)
}

// GetStream performs an HTTP GET request using provided client, URL and request headers.
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
		errPayload, err := io.ReadAll(res.Body)
		if err == nil && len(errPayload) > 0 {
			errorMsg += ": " + string(errPayload)
		}
		return nil, respContentType, fmt.Errorf(errorMsg)
	}

	return res.Body, respContentType, nil
}

// InitHTTPClient returns an HTTP client using the configuration from the apprepo and CA secret given.
func InitHTTPClient(appRepo *v1alpha1.AppRepository, caCertSecret *corev1.Secret) (*http.Client, error) {
	// create cert pool
	var certsData []byte = nil
	if caCertSecret != nil && appRepo.Spec.Auth.CustomCA != nil {
		// Fetch cert data
		key := appRepo.Spec.Auth.CustomCA.SecretKeyRef.Key
		customData, ok := caCertSecret.Data[key]
		if !ok {
			customDataString, ok := caCertSecret.StringData[key]
			if !ok {
				return nil, fmt.Errorf("secret %q did not contain key %q", appRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, key)
			}
			customData = []byte(customDataString)
		}
		certsData = customData
	}
	caCertPool, err := GetCertPool(certsData)
	if err != nil {
		return nil, err
	}

	// proxy config
	proxyConfig := getProxyConfig(appRepo)
	proxyFunc := func(r *http.Request) (*url.URL, error) { return proxyConfig.ProxyFunc()(r.URL) }

	// create client
	client := New()
	// #nosec G402
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: appRepo.Spec.TLSInsecureSkipVerify,
	}
	if err := SetClientTLS(client, tlsConfig); err != nil {
		return nil, err
	}
	if err := SetClientProxy(client, proxyFunc); err != nil {
		return nil, err
	}

	return client, nil
}

// InitNetClient returns an HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
func InitNetClient(appRepo *v1alpha1.AppRepository, caCertSecret, authSecret *corev1.Secret, defaultHeaders http.Header) (Client, error) {
	netClient, err := InitHTTPClient(appRepo, caCertSecret)
	if err != nil {
		return nil, err
	}

	if defaultHeaders == nil {
		defaultHeaders = http.Header{}
	}
	if authSecret != nil && appRepo.Spec.Auth.Header != nil {
		auth, err := kube.GetDataFromSecret(appRepo.Spec.Auth.Header.SecretKeyRef.Key, authSecret)
		if err != nil {
			return nil, err
		}
		defaultHeaders.Set("Authorization", auth)
	}

	return &ClientWithDefaults{
		Client:         netClient,
		DefaultHeaders: defaultHeaders,
	}, nil
}

func getProxyConfig(appRepo *v1alpha1.AppRepository) *httpproxy.Config {
	template := appRepo.Spec.SyncJobPodTemplate
	proxyConfig := httpproxy.Config{}
	defaultToEnv := true
	if len(template.Spec.Containers) > 0 {
		for _, e := range template.Spec.Containers[0].Env {
			switch e.Name {
			case "http_proxy":
				proxyConfig.HTTPProxy = e.Value
				defaultToEnv = false
			case "https_proxy":
				proxyConfig.HTTPSProxy = e.Value
				defaultToEnv = false
			case "no_proxy":
				proxyConfig.NoProxy = e.Value
				defaultToEnv = false
			}
		}
	}

	if defaultToEnv {
		return httpproxy.FromEnvironment()
	}

	return &proxyConfig
}
