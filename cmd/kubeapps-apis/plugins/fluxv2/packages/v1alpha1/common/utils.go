// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-redis/redis/v8"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"golang.org/x/net/http/httpproxy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// copied from helm plug-in
	UserAgentPrefix          = "kubeapps-apis/plugins"
	redisInitClientRetryWait = 1 * time.Second
	redisInitClientTimeout   = 10 * time.Second
)

// Set the pluginDetail once during a module init function so the single struct
// can be used throughout the plugin.
var (
	pluginDetail plugins.Plugin
	// This version var is updated during the build (see the -ldflags option
	// in the cmd/kubeapps-apis/Dockerfile)
	version         = "devel"
	repositoriesGvr schema.GroupVersionResource
	chartsGvr       schema.GroupVersionResource
	releasesGvr     schema.GroupVersionResource
)

func init() {
	pluginDetail = plugins.Plugin{
		Name:    "fluxv2.packages",
		Version: "v1alpha1",
	}

	repositoriesGvr = schema.GroupVersionResource{
		Group:    sourcev1.GroupVersion.Group,
		Version:  sourcev1.GroupVersion.Version,
		Resource: "helmrepositories",
	}

	chartsGvr = schema.GroupVersionResource{
		Group:    sourcev1.GroupVersion.Group,
		Version:  sourcev1.GroupVersion.Version,
		Resource: "helmcharts",
	}

	releasesGvr = schema.GroupVersionResource{
		Group:    helmv2.GroupVersion.Group,
		Version:  helmv2.GroupVersion.Version,
		Resource: "helmreleases",
	}
}

//
// miscellaneous utility funcs
//
func NewDefaultPluginConfig() *FluxPluginConfig {
	// If no config is provided, we default to the existing values for backwards
	// compatibility.
	return &FluxPluginConfig{
		VersionsInSummary:    pkgutils.GetDefaultVersionsInSummary(),
		TimeoutSeconds:       int32(-1),
		DefaultUpgradePolicy: pkgutils.UpgradePolicyNone,
		UserManagedSecrets:   false,
	}
}

func PrettyPrint(o interface{}) string {
	prettyBytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", o)
	}
	return string(prettyBytes)
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

// "Local" in the sense of no namespace is specified
func NewLocalOpaqueSecret(name string) *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{},
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

// https://github.com/vmware-tanzu/kubeapps/pull/3044#discussion_r662733334
// small preference for reading all config in the main.go
// (whether from env vars or cmd-line options) only in the one spot and passing
// explicitly to functions (so functions are less dependent on env state).
func NewRedisClientFromEnv(stopCh <-chan struct{}) (*redis.Client, error) {
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

	// ref https://github.com/vmware-tanzu/kubeapps/pull/4382#discussion_r820386531
	var redisCli *redis.Client
	err = wait.PollImmediate(redisInitClientRetryWait, redisInitClientTimeout,
		func() (bool, error) {
			redisCli = redis.NewClient(&redis.Options{
				Addr:     REDIS_ADDR,
				Password: REDIS_PASSWORD,
				DB:       REDIS_DB_NUM,
			})

			// ping redis to make sure client is connected
			var pong string
			if pong, err = redisCli.Ping(redisCli.Context()).Result(); err == nil {
				log.Infof("Redis [PING]: %s", pong)
				return true, nil
			}
			log.Infof("Waiting %s before retrying to due to %v...", redisInitClientRetryWait.String(), err)
			return false, nil
		})

	if err != nil {
		return nil, fmt.Errorf("initializing redis client failed after timeout of %s was reached, error: %v", redisInitClientTimeout.String(), err)
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

type FluxPluginConfig struct {
	VersionsInSummary    pkgutils.VersionsInSummary
	TimeoutSeconds       int32
	DefaultUpgradePolicy pkgutils.UpgradePolicy
	// whether or not secrets are fully managed by user or kubeapps
	// see comments in design spec under AddPackageRepository.
	// false (i.e. kubeapps manages secrets) by default
	UserManagedSecrets bool
}

// ParsePluginConfig parses the input plugin configuration json file and return the
// configuration options.
func ParsePluginConfig(pluginConfigPath string) (*FluxPluginConfig, error) {
	// In the flux plugin, for example, we are interested in
	// a) config for the core.packages.v1alpha1.
	// b) flux plugin-specific config
	type internalFluxPluginConfig struct {
		Core struct {
			Packages struct {
				V1alpha1 struct {
					VersionsInSummary pkgutils.VersionsInSummary
					TimeoutSeconds    int32 `json:"timeoutSeconds"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"core"`

		Flux struct {
			Packages struct {
				V1alpha1 struct {
					DefaultUpgradePolicy string `json:"defaultUpgradePolicy"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"flux"`
	}
	var config internalFluxPluginConfig

	pluginConfig, err := ioutil.ReadFile(pluginConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open plugin config at %q: %w", pluginConfigPath, err)
	}
	err = json.Unmarshal([]byte(pluginConfig), &config)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal plugin config: %q error: %w", string(pluginConfig), err)
	}

	if defaultUpgradePolicy, err := pkgutils.UpgradePolicyFromString(
		config.Flux.Packages.V1alpha1.DefaultUpgradePolicy); err != nil {
		return nil, err
	} else {
		// return configured value
		return &FluxPluginConfig{
			VersionsInSummary:    config.Core.Packages.V1alpha1.VersionsInSummary,
			TimeoutSeconds:       config.Core.Packages.V1alpha1.TimeoutSeconds,
			DefaultUpgradePolicy: defaultUpgradePolicy,
			UserManagedSecrets:   false,
		}, nil
	}
}

func GetRepositoriesGvr() schema.GroupVersionResource {
	return repositoriesGvr
}

func GetChartsGvr() schema.GroupVersionResource {
	return chartsGvr
}

func GetReleasesGvr() schema.GroupVersionResource {
	return releasesGvr
}

func DockerCredentialsToSecretData(dc *corev1.DockerCredentials) []byte {
	// ref https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
	authStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", dc.Username, dc.Password)))
	configStr := fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"email\":\"%s\",\"auth\":\"%s\"}}}",
		dc.Server, dc.Username, dc.Password, dc.Email, authStr)
	return []byte(base64.StdEncoding.EncodeToString([]byte(configStr)))
}

func SecretDataToDockerCredentials(configStr string) (*corev1.DockerCredentials, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(string(configStr))
	if err != nil {
		return nil, err
	}
	var configMap map[string]interface{}
	err = json.Unmarshal(decodedStr, &configMap)
	if err != nil {
		return nil, err
	}
	auths, ok := configMap["auths"]
	if !ok {
		return nil, status.Errorf(codes.Internal, "secret data is not in expected format")
	}
	authsMap, ok := auths.(map[string]interface{})
	if !ok {
		return nil, status.Errorf(codes.Internal, "secret data is not in expected format")
	}
	server := ""
	for k := range authsMap {
		server = k
		break
	}
	if server == "" {
		return nil, status.Errorf(codes.Internal, "secret data is not in expected format")
	}
	serverMap, ok := authsMap[server].(map[string]interface{})
	if !ok {
		return nil, status.Errorf(codes.Internal, "secret data is not in expected format")
	}
	username, ok := serverMap["username"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "secret data is not in expected format")
	}
	password, ok := serverMap["password"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "secret data is not in expected format")
	}
	email, ok := serverMap["email"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "secret data is not in expected format")
	}
	return &corev1.DockerCredentials{
		Server:   server,
		Username: username,
		Password: password,
		Email:    email,
	}, nil
}
