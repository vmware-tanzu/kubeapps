// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// configForCluster implements the genericclioptions.RESTClientGetter interface
// while ensuring that it returns the config it was given, rather
// than re-creating a config as if the CLI options were passed in as the
// genericclioptions.ConfigFlags implementation, which strips the bearer token
// if the host is not https (helpful for the CLI).
//
// TODO(mnelson) The better short-term option is to update pinniped-proxy to support TLS
// even though it's for internal traffic only, as it will be required in many circumstances.
// https://github.com/vmware-tanzu/kubeapps/issues/2268
// This implementation can be completely removed once TLS is used by pinniped-proxy.
type configForCluster struct {
	config         *rest.Config
	discoveryBurst int
	*genericclioptions.ConfigFlags
}

// ToRESTConfig has the main difference from the original implementation,
// returning the config as is.
func (f *configForCluster) ToRESTConfig() (*rest.Config, error) {
	return f.config, nil
}

// ToDiscoveryClient requires an implementation on the embedding struct because
// the implementation calls ToRESTConfig(). Painfully, this then requires copying
// the complete function.
func (f *configForCluster) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := f.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	// The more groups you have, the more discovery requests you need to make.
	// given 25 groups (our groups + a few custom resources) with one-ish version each, discovery needs to make 50 requests
	// double it just so we don't end up here again for a while.  This config is only used for discovery.
	config.Burst = f.discoveryBurst

	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	cacheDir := filepath.Join(home, ".kube", "cache")

	// retrieve a user-provided value for the "cache-dir"
	// override httpCacheDir and discoveryCacheDir if user-value is given.
	if f.CacheDir != nil {
		cacheDir = *f.CacheDir
	}
	httpCacheDir := filepath.Join(cacheDir, "http")
	discoveryCacheDir := computeDiscoverCacheDir(filepath.Join(cacheDir, "discovery"), config.Host)

	return diskcached.NewCachedDiscoveryClientForConfig(config, discoveryCacheDir, httpCacheDir, time.Duration(10*time.Minute))
}

// ToRESTMapper requires an implementation on the embedding struct because
// it calls ToDiscoveryClient.
func (f *configForCluster) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := f.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

// ToRawKubeConfigLoader may need to be replicated here, but from limited
// testing it works just to call the embedded implementation.
func (f *configForCluster) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return f.ConfigFlags.ToRawKubeConfigLoader()
}

// WithDiscoveryBurst is required because it sets the private var used by
// ToDiscoveryClient.
func (f *configForCluster) WithDiscoveryBurst(discoveryBurst int) *configForCluster {
	f.discoveryBurst = discoveryBurst
	return f
}

// The following var and func are only required here because computeDiscoverCacheDir
// is used in ToDiscoveryClient above.

// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/\.)]`)

// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if its there:
	schemelessHost := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	return filepath.Join(parentDir, safeHost)
}
