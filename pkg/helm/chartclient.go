// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"golang.org/x/net/http/httpproxy"
	corev1 "k8s.io/api/core/v1"
)

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
func InitNetClient(appRepo *v1alpha1.AppRepository, caCertSecret, authSecret *corev1.Secret, defaultHeaders http.Header) (*http.Client, error) {
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

	return httpclient.NewDefaultHeaderClient(netClient, defaultHeaders), nil
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
