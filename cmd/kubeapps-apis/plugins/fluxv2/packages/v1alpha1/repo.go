// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/helm"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// see docs at https://fluxcd.io/docs/components/source/ and
	// https://fluxcd.io/docs/components/helm/api/
	fluxHelmRepositories   = "helmrepositories"
	fluxHelmRepositoryList = "HelmRepositoryList"
)

var (
	// default poll interval is 10 min
	defaultPollInterval = metav1.Duration{Duration: 10 * time.Minute}
)

// namespace maybe apiv1.NamespaceAll, in which case repositories from all namespaces are returned
func (s *Server) listReposInNamespace(ctx context.Context, namespace string) ([]sourcev1.HelmRepository, error) {
	client, err := s.getClient(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var repoList sourcev1.HelmRepositoryList
	if err := client.List(ctx, &repoList); err != nil {
		return nil, statuserror.FromK8sError("list", "HelmRepository", namespace+"/*", err)
	} else {
		return repoList.Items, nil
	}
}

func (s *Server) getRepoInCluster(ctx context.Context, key types.NamespacedName) (*sourcev1.HelmRepository, error) {
	client, err := s.getClient(ctx, key.Namespace)
	if err != nil {
		return nil, err
	}
	var repo sourcev1.HelmRepository
	if err = client.Get(ctx, key, &repo); err != nil {
		return nil, statuserror.FromK8sError("get", "HelmRepository", key.String(), err)
	}
	return &repo, nil
}

// regexp expressions are used for matching actual names against expected patters
func (s *Server) filterReadyReposByName(repoList []sourcev1.HelmRepository, match []string) (sets.String, error) {
	if s.repoCache == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server cache has not been properly initialized")
	}

	resultKeys := sets.String{}
	for _, repo := range repoList {
		// first check if repo is in ready state
		if !isRepoReady(repo) {
			// just skip it
			continue
		}
		name, err := common.NamespacedName(&repo)
		if err != nil {
			// just skip it
			continue
		}
		// see if name matches the filter
		matched := false
		if len(match) > 0 {
			for _, m := range match {
				if matched, err = regexp.MatchString(m, name.Name); matched && err == nil {
					break
				}
			}
		} else {
			matched = true
		}
		if matched {
			resultKeys.Insert(s.repoCache.KeyForNamespacedName(*name))
		}
	}
	return resultKeys, nil
}

func (s *Server) getChartsForRepos(ctx context.Context, match []string) (map[string][]models.Chart, error) {
	// 1. with flux an available package may be from a repo in any namespace
	// 2. can't rely on cache as a real source of truth for key names
	//    because redis may evict cache entries due to memory pressure to make room for new ones
	repoList, err := s.listReposInNamespace(ctx, apiv1.NamespaceAll)
	if err != nil {
		return nil, err
	}

	repoNames, err := s.filterReadyReposByName(repoList, match)
	if err != nil {
		return nil, err
	}

	chartsUntyped, err := s.repoCache.GetForMultiple(repoNames)
	if err != nil {
		return nil, err
	}

	chartsTyped := make(map[string][]models.Chart)
	for key, value := range chartsUntyped {
		if value == nil {
			chartsTyped[key] = nil
		} else {
			typedValue, ok := value.(repoCacheEntryValue)
			if !ok {
				return nil, status.Errorf(
					codes.Internal,
					"unexpected value fetched from cache: type: [%s], value: [%v]",
					reflect.TypeOf(value), value)
			}
			chartsTyped[key] = typedValue.Charts
		}
	}
	return chartsTyped, nil
}

func (s *Server) clientOptionsForRepo(ctx context.Context, repoName types.NamespacedName) (*common.ClientOptions, error) {
	repo, err := s.getRepoInCluster(ctx, repoName)
	if err != nil {
		return nil, err
	}
	// notice a bit of inconsistency here, we are using s.clientGetter
	// (i.e. the context of the incoming request) to read the secret
	// as opposed to s.repoCache.clientGetter (which uses the context of
	//	User "system:serviceaccount:kubeapps:kubeapps-internal-kubeappsapis")
	// which is what is used when the repo is being processed/indexed.
	// I don't think it's necessarily a bad thing if the incoming user's RBAC
	// settings are more permissive than that of the default RBAC for
	// kubeapps-internal-kubeappsapis account. If we don't like that behavior,
	// I can easily switch to BackgroundClientGetter here
	sink := repoEventSink{
		clientGetter: s.newBackgroundClientGetter(),
		chartCache:   s.chartCache,
	}
	return sink.clientOptionsForRepo(ctx, *repo)
}

func (s *Server) newRepo(ctx context.Context, targetName types.NamespacedName, url string, interval uint32,
	tlsConfig *corev1.PackageRepositoryTlsConfig, auth *corev1.PackageRepositoryAuth) (*corev1.PackageRepositoryReference, error) {
	if url == "" {
		return nil, status.Errorf(codes.InvalidArgument, "repository url may not be empty")
	}

	var secret *apiv1.Secret
	if tlsConfig != nil {
		if tlsConfig.InsecureSkipVerify {
			return nil, status.Errorf(codes.Unimplemented, "TLS flag insecureSkipVerify is not supported")
		}
		caCert := tlsConfig.GetCertAuthority()
		if caCert != "" {
			secret = newLocalOpaqueSecret(targetName.Name + "-")
			secret.Data["caFile"] = []byte(caCert)
		}
	}
	if auth != nil && auth.GetSecretRef() == nil {
		if secret == nil {
			secret = newLocalOpaqueSecret(targetName.Name + "-")
		}
		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			if unp := auth.GetUsernamePassword(); unp != nil {
				secret.Data["username"] = []byte(unp.Username)
				secret.Data["password"] = []byte(unp.Password)
			} else {
				return nil, status.Errorf(codes.Internal, "Username/Password configuration is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
			if ck := auth.GetTlsCertKey(); ck != nil {
				secret.Data["certFile"] = []byte(ck.Cert)
				secret.Data["keyFile"] = []byte(ck.Key)
			} else {
				return nil, status.Errorf(codes.Internal, "TLS Cert/Key configuration is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER, corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM:
			return nil, status.Errorf(codes.Unimplemented, "Package repository authentication type %q is not supported", auth.Type)
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if dc := auth.GetDockerCreds(); dc != nil {
				secret.Type = apiv1.SecretTypeDockerConfigJson
				// ref https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
				authStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", dc.Username, dc.Password)))
				configStr := fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"email\":\"%s\",\"auth\":\"%s\"}}}",
					dc.Server, dc.Username, dc.Password, dc.Email, authStr)
				secret.Data[".dockerconfigjson"] = []byte(base64.StdEncoding.EncodeToString([]byte(configStr)))
			} else {
				return nil, status.Errorf(codes.Internal, "Docker credentials configuration is missing")
			}
		default:
			return nil, status.Errorf(codes.Internal, "Unexpected package repository authentication type: %q", auth.Type)
		}
	}

	passCredentials := auth != nil && auth.PassCredentials

	secretRef, checkSecret := "", false
	if secret != nil {
		// create a secret first, if applicable
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return nil, err
		} else if secret, err = typedClient.CoreV1().Secrets(targetName.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return nil, statuserror.FromK8sError("create", "secret", secret.GetName(), err)
		} else {
			secretRef = secret.GetName()
		}
	} else if tlsConfig != nil && tlsConfig.GetSecretRef().GetName() != "" {
		// check that the specified secret exists
		secretRef, checkSecret = tlsConfig.GetSecretRef().GetName(), true
	} else if auth != nil && auth.GetSecretRef().GetName() != "" {
		secretRef, checkSecret = auth.GetSecretRef().GetName(), true
	}

	if secretRef != "" && checkSecret {
		// check that the specified secret exists
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return nil, err
		} else if _, err = typedClient.CoreV1().Secrets(targetName.Namespace).Get(ctx, secretRef, metav1.GetOptions{}); err != nil {
			return nil, statuserror.FromK8sError("get", "secret", secretRef, err)
		}
		// TODO (gfichtenholt) also check that the data in the opaque secret corresponds
		// to specified auth type, e.g. if AuthType is
		// PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
		// check that the secret has "username" and "password" fields, etc.

		// TODO (gfichtenholt)
		// ref https://github.com/kubeapps/kubeapps/pull/4353#discussion_r816332595
		// check whether flux supports typed secrets in addition to opaque secrets
		// https://kubernetes.io/docs/concepts/configuration/secret/#secret-types
		// If so, that cause certain validation to be done on the data (ie. ensuring that
		//	the "username" and "password" fields are present).
	}

	if fluxRepo, err := s.newFluxHelmRepo(targetName, url, interval, secretRef, passCredentials); err != nil {
		return nil, err
	} else if client, err := s.getClient(ctx, targetName.Namespace); err != nil {
		return nil, err
	} else if err = client.Create(ctx, fluxRepo); err != nil {
		return nil, statuserror.FromK8sError("create", "HelmRepository", targetName.String(), err)
	} else {
		return &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Namespace: fluxRepo.Namespace,
				Cluster:   s.kubeappsCluster,
			},
			Identifier: fluxRepo.Name,
			Plugin:     GetPluginDetail(),
		}, nil
	}
}

// ref https://fluxcd.io/docs/components/source/helmrepositories/
func (s *Server) newFluxHelmRepo(
	targetName types.NamespacedName,
	url string,
	interval uint32,
	secretRef string,
	passCredentials bool) (*sourcev1.HelmRepository, error) {
	pollInterval := defaultPollInterval
	if interval > 0 {
		pollInterval = metav1.Duration{Duration: time.Duration(interval) * time.Second}
	}
	fluxRepo := &sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName.Name,
			Namespace: targetName.Namespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:      url,
			Interval: pollInterval,
		},
	}
	if secretRef != "" {
		fluxRepo.Spec.SecretRef = &fluxmeta.LocalObjectReference{
			Name: secretRef,
		}
	}
	if passCredentials {
		fluxRepo.Spec.PassCredentials = true
	}
	return fluxRepo, nil
}

//
// implements plug-in specific cache-related functionality
//
type repoEventSink struct {
	clientGetter clientgetter.BackgroundClientGetterFunc
	chartCache   *cache.ChartCache // chartCache maybe nil only in unit tests
}

// this is what we store in the cache for each cached repo
// all struct fields are capitalized so they're exported by gob encoding
type repoCacheEntryValue struct {
	Checksum string
	Charts   []models.Chart
}

// onAddRepo essentially tells the cache whether to and what to store for a given key
func (s *repoEventSink) onAddRepo(key string, obj ctrlclient.Object) (interface{}, bool, error) {
	log.V(4).Info("+onAddRepo(%s)", key)
	defer log.V(4).Info("-onAddRepo()")

	if repo, ok := obj.(*sourcev1.HelmRepository); !ok {
		return nil, false, fmt.Errorf("expected an instance of *sourcev1.HelmRepository, got: %s", reflect.TypeOf(obj))
	} else if isRepoReady(*repo) {
		// first, check the repo is ready
		// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
		if artifact := repo.GetArtifact(); artifact != nil {
			if checksum := artifact.Checksum; checksum == "" {
				return nil, false, status.Errorf(codes.Internal,
					"expected field status.artifact.checksum not found on HelmRepository\n[%s]",
					common.PrettyPrint(repo))
			} else {
				return s.indexAndEncode(checksum, *repo)
			}
		} else {
			return nil, false, status.Errorf(codes.Internal,
				"expected field status.artifact not found on HelmRepository\n[%s]",
				common.PrettyPrint(repo))
		}
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

func (s *repoEventSink) indexAndEncode(checksum string, repo sourcev1.HelmRepository) ([]byte, bool, error) {
	charts, err := s.indexOneRepo(repo)
	if err != nil {
		return nil, false, err
	}

	cacheEntryValue := repoCacheEntryValue{
		Checksum: checksum,
		Charts:   charts,
	}

	// use gob encoding instead of json, it peforms much better
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(cacheEntryValue); err != nil {
		return nil, false, err
	}

	if s.chartCache != nil {
		if opts, err := s.clientOptionsForRepo(context.Background(), repo); err != nil {
			// ref: https://github.com/kubeapps/kubeapps/pull/3899#issuecomment-990446931
			// I don't want this func to fail onAdd/onModify() if we can't read
			// the corresponding secret due to something like default RBAC settings:
			// "secrets "podinfo-basic-auth-secret" is forbidden:
			// User "system:serviceaccount:kubeapps:kubeapps-internal-kubeappsapis" cannot get
			// resource "secrets" in API group "" in the namespace "default"
			// So we still finish the indexing of the repo but skip the charts
			log.Errorf("Failed to read secret for repo due to: %+v", err)
		} else if err = s.chartCache.SyncCharts(charts, opts); err != nil {
			return nil, false, err
		}
	}
	return buf.Bytes(), true, nil
}

// it is assumed the caller has already checked that this repo is ready
// At present, there is only one caller of indexOneRepo() and this check is already done by it
func (s *repoEventSink) indexOneRepo(repo sourcev1.HelmRepository) ([]models.Chart, error) {
	startTime := time.Now()

	// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
	indexUrl := repo.Status.URL
	if indexUrl == "" {
		return nil, status.Errorf(codes.Internal,
			"expected field status.url not found on HelmRepository\n[%s]",
			repo.Name)
	}

	log.Infof("+indexOneRepo: [%s], index URL: [%s]", repo.Name, indexUrl)

	// In production, there should be no need to provide authz, userAgent or any of the TLS details,
	// as we are reading index.yaml file from local cluster, not some remote repo.
	// e.g. http://source-controller.flux-system.svc.cluster.local./helmrepository/default/bitnami/index.yaml
	// Flux does the hard work of pulling the index file from remote repo
	// into local cluster based on secretRef associated with HelmRepository, if applicable
	// This is only true of index.yaml, not the individual chart URLs within it

	// if a transient error occurs the item will be re-queued and retried after a back-off period
	byteArray, err := httpclient.Get(indexUrl, httpclient.New(), nil)
	if err != nil {
		return nil, err
	}

	modelRepo := &models.Repo{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		URL:       indexUrl,
		Type:      "helm",
	}

	// this is potentially a very expensive operation for large repos like 'bitnami'
	// shallow = true  => 8-9 sec
	// shallow = false => 12-13 sec, so deep copy adds 50% to cost, but we need it to
	// for GetAvailablePackageVersions()
	charts, err := helm.ChartsFromIndex(byteArray, modelRepo, false)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	msg := fmt.Sprintf("-indexOneRepo: [%s], indexed [%d] packages in [%d] ms", repo.Name, len(charts), duration.Milliseconds())
	if len(charts) > 0 {
		log.Info(msg)
	} else {
		// this is kind of a red flag - an index with 0 charts, most likely contents of index.yaml is
		// messed up and didn't parse successfully but the helm library didn't raise an error
		log.Warning(msg)
	}
	return charts, nil
}

// onModifyRepo essentially tells the cache whether or not to and what to store for a given key
func (s *repoEventSink) onModifyRepo(key string, obj ctrlclient.Object, oldValue interface{}) (interface{}, bool, error) {
	if repo, ok := obj.(*sourcev1.HelmRepository); !ok {
		return nil, false, fmt.Errorf("expected an instance of *sourcev1.HelmRepository, got: %s", reflect.TypeOf(obj))
	} else if isRepoReady(*repo) {
		// first check the repo is ready
		// We should to compare checksums on what's stored in the cache
		// vs the modified object to see if the contents has really changed before embarking on
		// expensive operation indexOneRepo() below.
		// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
		// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
		var newChecksum string
		if artifact := repo.GetArtifact(); artifact != nil {
			if newChecksum = artifact.Checksum; newChecksum == "" {
				return nil, false, status.Errorf(codes.Internal,
					"expected field status.artifact.checksum not found on HelmRepository\n[%s]",
					common.PrettyPrint(repo))
			}
		} else {
			return nil, false, status.Errorf(codes.Internal,
				"expected field status.artifact not found on HelmRepository\n[%s]",
				common.PrettyPrint(repo))
		}

		cacheEntryUntyped, err := s.onGetRepo(key, oldValue)
		if err != nil {
			return nil, false, err
		}

		cacheEntry, ok := cacheEntryUntyped.(repoCacheEntryValue)
		if !ok {
			return nil, false, status.Errorf(
				codes.Internal,
				"unexpected value found in cache for key [%s]: %v",
				key, cacheEntryUntyped)
		}

		if cacheEntry.Checksum != newChecksum {
			return s.indexAndEncode(newChecksum, *repo)
		} else {
			// skip because the content did not change
			return nil, false, nil
		}
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.V(4).Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

func (s *repoEventSink) onGetRepo(key string, value interface{}) (interface{}, error) {
	b, ok := value.([]byte)
	if !ok {
		return nil, status.Errorf(codes.Internal, "unexpected value found in cache for key [%s]: %v", key, value)
	}

	dec := gob.NewDecoder(bytes.NewReader(b))
	var entryValue repoCacheEntryValue
	if err := dec.Decode(&entryValue); err != nil {
		return nil, err
	}
	return entryValue, nil
}

func (s *repoEventSink) onDeleteRepo(key string) (bool, error) {
	if s.chartCache != nil {
		if name, err := s.fromKey(key); err != nil {
			return false, err
		} else if err := s.chartCache.DeleteChartsForRepo(name); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *repoEventSink) onResync() error {
	if s.chartCache != nil {
		return s.chartCache.OnResync()
	} else {
		return nil
	}
}

// TODO (gfichtenholt) low priority: don't really like the fact that these 4 lines of code
// basically repeat same logic as NamespacedResourceWatcherCache.fromKey() but can't
// quite come up with with a more elegant alternative right now
func (s *repoEventSink) fromKey(key string) (*types.NamespacedName, error) {
	parts := strings.Split(key, cache.KeySegmentsSeparator)
	if len(parts) != 3 || parts[0] != fluxHelmRepositories || len(parts[1]) == 0 || len(parts[2]) == 0 {
		return nil, status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	return &types.NamespacedName{Namespace: parts[1], Name: parts[2]}, nil
}

// this is only until https://github.com/kubeapps/kubeapps/issues/3496
// "Investigate and propose package repositories API with similar core interface to packages API"
// gets implemented. After that, the auth should be part of some kind of packageRepositoryFromCtrlObject()
// The reason I do this here is to set up auth that may be needed to fetch chart tarballs by
// ChartCache
func (s *repoEventSink) clientOptionsForRepo(ctx context.Context, repo sourcev1.HelmRepository) (*common.ClientOptions, error) {
	if repo.Spec.SecretRef == nil {
		return nil, nil
	}
	secretName := repo.Spec.SecretRef.Name
	if secretName == "" {
		return nil, nil
	}
	if s == nil || s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "unexpected state in clientGetterHolder instance")
	}
	typedClient, err := s.clientGetter.Typed(ctx)
	if err != nil {
		return nil, err
	}
	repoName, err := common.NamespacedName(&repo)
	if err != nil {
		return nil, err
	}
	secret, err := typedClient.CoreV1().Secrets(repoName.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("get", "secret", secretName, err)
	}
	return common.ClientOptionsFromSecret(*secret)
}

//
// repo-related utilities
//

func isRepoReady(repo sourcev1.HelmRepository) bool {
	// see docs at https://fluxcd.io/docs/components/source/helmrepositories/
	// Confirm the state we are observing is for the current generation
	if !common.CheckGeneration(&repo) {
		return false
	}

	completed, success, _ := isHelmRepositoryReady(repo)
	return completed && success
}

// returns 3 things:
// - complete whether the operation was completed
// - success (only applicable when complete == true) whether the operation was successful or failed
// - reason, if present
// docs:
// 1. https://fluxcd.io/docs/components/source/helmrepositories/#status-examples
func isHelmRepositoryReady(repo sourcev1.HelmRepository) (complete bool, success bool, reason string) {
	if !common.CheckGeneration(&repo) {
		return false, false, ""
	}

	readyCond := meta.FindStatusCondition(*repo.GetStatusConditions(), "Ready")
	if readyCond != nil {
		if readyCond.Reason != "" {
			// this could be something like
			// "reason": "ChartPullFailed"
			// i.e. not super-useful
			reason = readyCond.Reason
		}
		if readyCond.Message != "" {
			// whereas this could be something like:
			// "message": 'invalid chart URL format'
			// i.e. a little more useful, so we'll just return them both
			reason += ": " + readyCond.Message
		}
		switch readyCond.Status {
		case metav1.ConditionTrue:
			return true, true, reason
		case metav1.ConditionFalse:
			return true, false, reason
			// metav1.ConditionUnknown falls through
		}
	}
	return false, false, reason
}

func newLocalOpaqueSecret(name string) *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name,
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}
}
