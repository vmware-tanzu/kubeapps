/*
Copyright (c) 2020 Bitnami

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

package kube

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"golang.org/x/net/http/httpproxy"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"
)

const (
	defaultTimeoutSeconds = 180
)

// HTTPClient Interface to perform HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// clientWithDefaultHeaders implements chart.HTTPClient interface
// and includes an override of the Do method which injects our default
// headers - User-Agent and Authorization (when present)
type clientWithDefaultHeaders struct {
	client         HTTPClient
	defaultHeaders http.Header
}

// Do HTTP request
func (c *clientWithDefaultHeaders) Do(req *http.Request) (*http.Response, error) {
	for k, v := range c.defaultHeaders {
		// Only add the default header if it's not already set in the request.
		if _, ok := req.Header[k]; !ok {
			req.Header[k] = v
		}
	}
	return c.client.Do(req)
}

// getDataFromRegistrySecret retrieves the given key from the secret as a string
func getDataFromRegistrySecret(key string, s *corev1.Secret) (string, error) {
	dockerConfigJsonEncoded, ok := s.StringData[key]
	if !ok {
		authBytes, ok := s.Data[key]
		if !ok {
			return "", fmt.Errorf("secret %q did not contain key %q", s.Name, key)
		}
		dockerConfigJsonEncoded = string(authBytes)
	}
	dockerConfigJson, err := base64.StdEncoding.DecodeString(dockerConfigJsonEncoded)
	if err != nil {
		return "", fmt.Errorf("Unable to decode docker config secret. Got: %v", err)
	}

	dockerConfig := &credentialprovider.DockerConfigJson{}
	err = json.Unmarshal(dockerConfigJson, dockerConfig)
	if err != nil {
		return "", fmt.Errorf("Unable to parse secret %s as a Docker config. Got: %v", s.Name, err)
	}

	// This is a simplified handler of a Docker config which only looks for the username:password
	// of the first entry.
	for _, entry := range dockerConfig.Auths {
		auth := fmt.Sprintf("%s:%s", entry.Username, entry.Password)
		authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
		return authHeader, nil
	}

	return "", fmt.Errorf("secret %q did not contain docker creds", s.Name)
}

// GetData retrieves the given key from the secret as a string
func GetData(key string, s *corev1.Secret) (string, error) {
	if key == ".dockerconfigjson" {
		// Parse the secret as a docker registry secret
		return getDataFromRegistrySecret(key, s)
	}
	// Parse the secret as a plain secret
	auth, ok := s.StringData[key]
	if !ok {
		authBytes, ok := s.Data[key]
		if !ok {
			return "", fmt.Errorf("secret %q did not contain key %q", s.Name, key)
		}
		auth = string(authBytes)
	}
	return auth, nil
}

// InitHTTPClient returns a HTTP client using the configuration from the apprepo and CA secret given.
func InitHTTPClient(appRepo *v1alpha1.AppRepository, caCertSecret *corev1.Secret) (*http.Client, error) {
	// Require the SystemCertPool unless the env var is explicitly set.
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		if _, ok := os.LookupEnv("TILLER_PROXY_ALLOW_EMPTY_CERT_POOL"); !ok {
			return nil, err
		}
		caCertPool = x509.NewCertPool()
	}

	if caCertSecret != nil && appRepo.Spec.Auth.CustomCA != nil {
		// Append our cert to the system pool
		key := appRepo.Spec.Auth.CustomCA.SecretKeyRef.Key
		customData, ok := caCertSecret.Data[key]
		if !ok {
			customDataString, ok := caCertSecret.StringData[key]
			if !ok {
				return nil, fmt.Errorf("secret %q did not contain key %q", appRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, key)
			}
			customData = []byte(customDataString)
		}
		if ok := caCertPool.AppendCertsFromPEM(customData); !ok {
			return nil, fmt.Errorf("Failed to append %s to RootCAs", appRepo.Spec.Auth.CustomCA.SecretKeyRef.Name)
		}
	}
	proxyConfig := getProxyConfig(appRepo)
	proxyFunc := func(r *http.Request) (*url.URL, error) { return proxyConfig.ProxyFunc()(r.URL) }

	return &http.Client{
		Timeout: time.Second * defaultTimeoutSeconds,
		Transport: &http.Transport{
			Proxy: proxyFunc,
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: appRepo.Spec.TLSInsecureSkipVerify,
			},
		},
	}, nil
}

// InitNetClient returns an HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
func InitNetClient(appRepo *v1alpha1.AppRepository, caCertSecret, authSecret *corev1.Secret, defaultHeaders http.Header) (HTTPClient, error) {
	netClient, err := InitHTTPClient(appRepo, caCertSecret)
	if err != nil {
		return nil, err
	}

	if defaultHeaders == nil {
		defaultHeaders = http.Header{}
	}
	if authSecret != nil && appRepo.Spec.Auth.Header != nil {
		auth, err := GetData(appRepo.Spec.Auth.Header.SecretKeyRef.Key, authSecret)
		if err != nil {
			return nil, err
		}
		defaultHeaders.Set("Authorization", string(auth))
	}

	return &clientWithDefaultHeaders{
		client:         netClient,
		defaultHeaders: defaultHeaders,
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
