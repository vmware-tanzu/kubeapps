// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"golang.org/x/net/http/httpproxy"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"
)

func GetAuthHeaderFromDockerConfig(dockerConfig *credentialprovider.DockerConfigJSON) (string, error) {
	if len(dockerConfig.Auths) > 1 {
		return "", fmt.Errorf("The given config should include one auth entry")
	}
	// This is a simplified handler of a Docker config which only looks for the username:password
	// of the first entry.
	for _, entry := range dockerConfig.Auths {
		auth := fmt.Sprintf("%s:%s", entry.Username, entry.Password)
		authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
		return authHeader, nil
	}
	return "", fmt.Errorf("The given config doesn't include an auth entry")
}

// getDataFromRegistrySecret retrieves the given key from the secret as a string
func getDataFromRegistrySecret(key string, s *corev1.Secret) (string, error) {
	dockerConfigJson, ok := s.Data[key]
	if !ok {
		return "", fmt.Errorf("secret %q did not contain key %q", s.Name, key)
	}

	dockerConfig := &credentialprovider.DockerConfigJSON{}
	err := json.Unmarshal(dockerConfigJson, dockerConfig)
	if err != nil {
		return "", fmt.Errorf("Unable to parse secret %s as a Docker config. Got: %v", s.Name, err)
	}

	return GetAuthHeaderFromDockerConfig(dockerConfig)
}

// GetDataFromSecret retrieves the given key from the secret as a string
func GetDataFromSecret(key string, s *corev1.Secret) (string, error) {
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
	caCertPool, err := httpclient.GetCertPool(certsData)
	if err != nil {
		return nil, err
	}

	// proxy config
	proxyConfig := getProxyConfig(appRepo)
	proxyFunc := func(r *http.Request) (*url.URL, error) { return proxyConfig.ProxyFunc()(r.URL) }

	// create client
	client := httpclient.New()
	// #nosec G402
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: appRepo.Spec.TLSInsecureSkipVerify,
	}
	if err := httpclient.SetClientTLS(client, tlsConfig); err != nil {
		return nil, err
	}
	if err := httpclient.SetClientProxy(client, proxyFunc); err != nil {
		return nil, err
	}

	return client, nil
}

// InitNetClient returns an HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
func InitNetClient(appRepo *v1alpha1.AppRepository, caCertSecret, authSecret *corev1.Secret, defaultHeaders http.Header) (httpclient.Client, error) {
	netClient, err := InitHTTPClient(appRepo, caCertSecret)
	if err != nil {
		return nil, err
	}

	if defaultHeaders == nil {
		defaultHeaders = http.Header{}
	}
	if authSecret != nil && appRepo.Spec.Auth.Header != nil {
		auth, err := GetDataFromSecret(appRepo.Spec.Auth.Header.SecretKeyRef.Key, authSecret)
		if err != nil {
			return nil, err
		}
		defaultHeaders.Set("Authorization", string(auth))
	}

	return &httpclient.ClientWithDefaults{
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
