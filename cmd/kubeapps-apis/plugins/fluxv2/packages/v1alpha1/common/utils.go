// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-redis/redis/v8"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"golang.org/x/net/http/httpproxy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// copied from helm plug-in
	UserAgentPrefix = "kubeapps-apis/plugins"
)

// Set the pluginDetail once during a module init function so the single struct
// can be used throughout the plugin.
var (
	pluginDetail plugins.Plugin
	// This version var is updated during the build (see the -ldflags option
	// in the cmd/kubeapps-apis/Dockerfile)
	version = "devel"
)

func init() {
	pluginDetail = plugins.Plugin{
		Name:    "fluxv2.packages",
		Version: "v1alpha1",
	}
}

//
// miscellaneous utility funcs
//
func PrettyPrint(o interface{}) string {
	prettyBytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", o)
	}
	return string(prettyBytes)
}

func CheckGeneration(obj ctrlclient.Object) bool {
	generation := obj.GetGeneration()
	var observedGeneration int64
	if repo, ok := obj.(*sourcev1.HelmRepository); ok {
		observedGeneration = repo.Status.ObservedGeneration
	} else if rel, ok := obj.(*helmv2.HelmRelease); ok {
		observedGeneration = rel.Status.ObservedGeneration
	} else {
		log.Errorf("Unsupported object type in CheckGeneration: %#v", obj)
		return false
	}
	return generation == observedGeneration
}

func NamespacedName(obj ctrlclient.Object) (*types.NamespacedName, error) {
	name := obj.GetName()
	namespace := obj.GetNamespace()
	if name != "" && namespace != "" {
		return &types.NamespacedName{Name: name, Namespace: namespace}, nil
	} else {
		return nil,
			status.Errorf(codes.Internal,
				"required fields 'metadata.name' and/or 'metadata.namespace' not found on resource: %v",
				PrettyPrint(obj))
	}
}

// ref: https://blog.trailofbits.com/2020/06/09/how-to-check-if-a-mutex-is-locked-in-go/
// I understand this is not really "kosher" in general for production usage,
// but in one specific case (cache populateWith() func) it's okay as a confidence test
// if it turns out not, I can always remove this check, it's not critical
const mutexLocked = 1

func RWMutexWriteLocked(rw *sync.RWMutex) bool {
	// RWMutex has a "w" sync.Mutex field for write lock
	state := reflect.ValueOf(rw).Elem().FieldByName("w").FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

// note this implementation not correct for all cases. Thank you @minelson.
// When there are active readers, and there is a concurrent .Lock() request for writing,
// the readerCount may become < 0.
// see https://github.com/golang/go/blob/release-branch.go1.14/src/sync/rwmutex.go#L100
// so this code definitely needs be used with caution or better avoided
func RWMutexReadLocked(rw *sync.RWMutex) bool {
	return reflect.ValueOf(rw).Elem().FieldByName("readerCount").Int() > 0
}

// https://github.com/kubeapps/kubeapps/pull/3044#discussion_r662733334
// small preference for reading all config in the main.go
// (whether from env vars or cmd-line options) only in the one spot and passing
// explicitly to functions (so functions are less dependent on env state).
func NewRedisClientFromEnv() (*redis.Client, error) {
	REDIS_ADDR, ok := os.LookupEnv("REDIS_ADDR")
	if !ok {
		return nil, fmt.Errorf("missing environment variable REDIS_ADDR")
	}
	REDIS_PASSWORD, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		return nil, fmt.Errorf("missing environment variable REDIS_PASSWORD")
	}
	REDIS_DB, ok := os.LookupEnv("REDIS_DB")
	if !ok {
		return nil, fmt.Errorf("missing environment variable REDIS_DB")
	}

	REDIS_DB_NUM, err := strconv.Atoi(REDIS_DB)
	if err != nil {
		return nil, err
	}

	redisCli := redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: REDIS_PASSWORD,
		DB:       REDIS_DB_NUM,
	})

	// confidence test that the redis client is connected
	if pong, err := redisCli.Ping(redisCli.Context()).Result(); err != nil {
		return nil, err
	} else {
		log.Infof("Redis [PING]: %s", pong)
	}

	if maxmemory, err := redisCli.ConfigGet(redisCli.Context(), "maxmemory").Result(); err != nil {
		return nil, err
	} else if len(maxmemory) > 1 {
		log.Infof("Redis [CONFIG GET maxmemory]: %v", maxmemory[1])
	}

	return redisCli, nil
}

func RedisMemoryStats(redisCli *redis.Client) (used, total string) {
	used, total = "?", "?"
	// ref: https://redis.io/commands/info
	if meminfo, err := redisCli.Info(redisCli.Context(), "memory").Result(); err == nil {
		for _, l := range strings.Split(meminfo, "\r\n") {
			if used == "?" && strings.HasPrefix(l, "used_memory_rss_human:") {
				used = strings.Split(l, ":")[1]
			} else if total == "?" && strings.HasPrefix(l, "maxmemory_human:") {
				total = strings.Split(l, ":")[1]
			}
			if used != "?" && total != "?" {
				break
			}
		}
	} else {
		log.Warningf("Failed to get redis memory stats due to: %v", err)
	}
	return used, total
}

// options are generic parameters to be provided to the httpclient during instantiation.
type ClientOptions struct {
	// for TLS connections
	CertBytes []byte
	KeyBytes  []byte
	CaBytes   []byte
	// for Basic Authentication
	Username string
	Password string
	// User Agent String:
	// "kubeapps-apis/plugins/fluxv2.packages/v1alpha1/devel"
	UserAgent string
}

// inspired by https://github.com/fluxcd/source-controller/blob/main/internal/helm/getter/getter.go#L29

// ClientOptionsFromSecret constructs a getter.Option slice for the given secret.
// It returns the slice, or an error.
func ClientOptionsFromSecret(secret apiv1.Secret) (*ClientOptions, error) {
	var opts ClientOptions
	if err := basicAuthFromSecret(secret, &opts); err != nil {
		return nil, err
	}
	if err := tlsClientConfigFromSecret(secret, &opts); err != nil {
		return nil, err
	}
	return &opts, nil
}

//
// Secrets with no username AND password are ignored, if only one is defined it
// returns an error.
func basicAuthFromSecret(secret apiv1.Secret, options *ClientOptions) error {
	username, password := string(secret.Data["username"]), string(secret.Data["password"])
	switch {
	case username == "" && password == "":
		return nil
	case username == "" || password == "":
		return fmt.Errorf("invalid '%s' secret data: required fields 'username' and 'password'", secret.Name)
	}
	options.Username = username
	options.Password = password
	return nil
}

// Secrets with no certFile, keyFile, AND caFile are ignored, if only a
// certBytes OR keyBytes is defined it returns an error.
func tlsClientConfigFromSecret(secret apiv1.Secret, options *ClientOptions) error {
	certBytes, keyBytes, caBytes := secret.Data["certFile"], secret.Data["keyFile"], secret.Data["caFile"]
	switch {
	case len(certBytes)+len(keyBytes)+len(caBytes) == 0:
		return nil
	case (len(certBytes) > 0 && len(keyBytes) == 0) || (len(keyBytes) > 0 && len(certBytes) == 0):
		return fmt.Errorf("invalid '%s' secret data: fields 'certFile' and 'keyFile' require each other's presence",
			secret.Name)
	}

	options.CaBytes = caBytes
	options.CertBytes = certBytes
	options.KeyBytes = keyBytes
	return nil
}

func NewHttpClientAndHeaders(clientOptions *ClientOptions) (*http.Client, map[string]string, error) {
	// I wish I could have re-used the code in pkg/chart/chart.go and pkg/kube_utils/kube_utils.go
	// InitHTTPClient(), etc. but alas, it's all built around AppRepository CRD, which I don't have.
	headers := make(map[string]string)
	headers["User-Agent"] = userAgentString()
	if clientOptions != nil {
		if clientOptions.Username != "" && clientOptions.Password != "" {
			auth := clientOptions.Username + ":" + clientOptions.Password
			headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		}
	}
	// In theory, the work queue should be able to retry transient errors
	// so I shouldn't have to do retries here. See above comment for explanation
	client := httpclient.New()
	if clientOptions != nil {
		if len(clientOptions.CaBytes) != 0 ||
			len(clientOptions.CertBytes) != 0 ||
			len(clientOptions.KeyBytes) != 0 {
			tlsConfig, err := httpclient.NewClientTLS(
				clientOptions.CertBytes, clientOptions.KeyBytes, clientOptions.CaBytes)
			if err != nil {
				return nil, nil, err
			} else {
				if err = httpclient.SetClientTLS(client, tlsConfig); err != nil {
					return nil, nil, err
				}
			}
		}
	}

	// proxy config
	proxyConfig := httpproxy.FromEnvironment()
	proxyFunc := func(r *http.Request) (*url.URL, error) { return proxyConfig.ProxyFunc()(r.URL) }
	if err := httpclient.SetClientProxy(client, proxyFunc); err != nil {
		return nil, nil, err
	}
	return client, headers, nil
}

// this string is the same for all outbound calls
func userAgentString() string {
	return fmt.Sprintf("%s/%s/%s/%s", UserAgentPrefix, pluginDetail.Name, pluginDetail.Version, version)
}

// GetPluginDetail returns a core.plugins.Plugin describing itself.
func GetPluginDetail() *plugins.Plugin {
	return &pluginDetail
}
