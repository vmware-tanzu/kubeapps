// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
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

// returns a list of HelmRepositories from all namespaces (cluster-wide), excluding
// the ones that the caller has no read access to
func (s *Server) listReposInAllNamespaces(ctx context.Context) ([]sourcev1.HelmRepository, error) {
	// the actual List(...) call will be executed in the context of
	// kubeapps-internal-kubeappsapis service account
	// ref https://github.com/vmware-tanzu/kubeapps/issues/4390 for explanation
	backgroundCtx := context.Background()
	client, err := s.serviceAccountClientGetter.ControllerRuntime(backgroundCtx)
	if err != nil {
		return nil, err
	}

	var repoList sourcev1.HelmRepositoryList
	if err := client.List(backgroundCtx, &repoList); err != nil {
		return nil, statuserror.FromK8sError("list", "HelmRepository", "", err)
	} else {
		// filter out those repos the caller has no access to
		namespaces := sets.String{}
		for _, item := range repoList.Items {
			namespaces.Insert(item.GetNamespace())
		}
		allowedNamespaces := sets.String{}
		gvr := common.GetRepositoriesGvr()
		for ns := range namespaces {
			if ok, err := s.hasAccessToNamespace(ctx, gvr, ns); err == nil && ok {
				allowedNamespaces.Insert(ns)
			} else if err != nil {
				return nil, err
			}
		}
		items := []sourcev1.HelmRepository{}
		for _, item := range repoList.Items {
			if allowedNamespaces.Has(item.GetNamespace()) {
				items = append(items, item)
			}
		}
		return items, nil
	}
}

func (s *Server) getRepoInCluster(ctx context.Context, key types.NamespacedName) (*sourcev1.HelmRepository, error) {
	// unlike List(), there is no need to execute Get() in the context of
	// kubeapps-internal-kubeappsapis service account and then filter out results based on
	// whether or not the caller hasAccessToNamespace(). We can just pass the caller
	// context into Get() and if the caller isn't allowed, Get will raise an error, which is what we
	// want
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

// Notes:
// 1. with flux, an available package may be from a repo in any namespace accessible to the caller
// 2. can't rely on cache as a real source of truth for key names
//    because redis may evict cache entries due to memory pressure to make room for new ones
func (s *Server) getChartsForRepos(ctx context.Context, match []string) (map[string][]models.Chart, error) {
	repoList, err := s.listReposInAllNamespaces(ctx)
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
	} else if tlsConfig != nil && tlsConfig.InsecureSkipVerify {
		return nil, status.Errorf(codes.InvalidArgument, "TLS flag insecureSkipVerify is not supported")
	}

	var secretRef string
	var err error
	if s.pluginConfig.UserManagedSecrets {
		if secretRef, err = s.validateRepoUserManagedSecrets(ctx, targetName, tlsConfig, auth); err != nil {
			return nil, err
		}
	} else {
		if secretRef, err = s.createRepoKubeappsManagedSecrets(ctx, targetName, tlsConfig, auth); err != nil {
			return nil, err
		}
	}

	passCredentials := auth != nil && auth.PassCredentials

	if fluxRepo, err := newFluxHelmRepo(targetName, url, interval, secretRef, passCredentials); err != nil {
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

func (s *Server) repoDetail(ctx context.Context, repoRef *corev1.PackageRepositoryReference) (*corev1.PackageRepositoryDetail, error) {
	key := types.NamespacedName{Namespace: repoRef.Context.Namespace, Name: repoRef.Identifier}

	repo, err := s.getRepoInCluster(ctx, key)
	if err != nil {
		return nil, err
	}

	var tlsConfig *corev1.PackageRepositoryTlsConfig
	var auth *corev1.PackageRepositoryAuth
	if repo.Spec.SecretRef != nil {
		secretName := repo.Spec.SecretRef.Name
		if s == nil || s.clientGetter == nil {
			return nil, status.Errorf(codes.Internal, "unexpected state in clientGetterHolder instance")
		}
		typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
		if err != nil {
			return nil, err
		}
		secret, err := typedClient.CoreV1().Secrets(repo.Namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, statuserror.FromK8sError("get", "secret", secretName, err)
		}

		if s.pluginConfig.UserManagedSecrets {
			if tlsConfig, auth, err = getRepoTlsConfigAndAuthWithUserManagedSecrets(secret); err != nil {
				return nil, err
			}
		} else {
			if tlsConfig, auth, err = getRepoTlsConfigAndAuthWithKubeappsManagedSecrets(secret); err != nil {
				return nil, err
			}
		}
	} else {
		auth = &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
		}
	}
	auth.PassCredentials = repo.Spec.PassCredentials
	return &corev1.PackageRepositoryDetail{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Namespace: repo.Namespace,
				Cluster:   s.kubeappsCluster,
			},
			Identifier: repo.Name,
			Plugin:     GetPluginDetail(),
		},
		Name: repo.Name,
		// TBD Flux HelmRepository CR doesn't have a designated field for description
		Description:     "",
		NamespaceScoped: false,
		Type:            "helm",
		Url:             repo.Spec.URL,
		Interval:        uint32(repo.Spec.Interval.Duration.Seconds()),
		TlsConfig:       tlsConfig,
		Auth:            auth,
		CustomDetail:    nil,
		Status:          repoStatus(*repo),
	}, nil
}

func (s *Server) repoSummaries(ctx context.Context, namespace string) ([]*corev1.PackageRepositorySummary, error) {
	summaries := []*corev1.PackageRepositorySummary{}
	var repos []sourcev1.HelmRepository
	var err error
	if namespace == apiv1.NamespaceAll {
		if repos, err = s.listReposInAllNamespaces(ctx); err != nil {
			return nil, err
		}
	} else {
		// here, the right semantics are different than that of availablePackageSummaries()
		// namely, if a specific namespace is passed in, we need to list repos in that namespace
		// and if the caller happens not to have 'read' access to that namespace, a PermissionDenied
		// error should be raised, as opposed to returning an empty list with no error
		var repoList sourcev1.HelmRepositoryList
		var client ctrlclient.Client
		if client, err = s.getClient(ctx, namespace); err != nil {
			return nil, err
		} else if err = client.List(ctx, &repoList); err != nil {
			return nil, statuserror.FromK8sError("list", "HelmRepository", "", err)
		} else {
			repos = repoList.Items
		}
	}
	for _, repo := range repos {
		summary := &corev1.PackageRepositorySummary{
			PackageRepoRef: &corev1.PackageRepositoryReference{
				Context: &corev1.Context{
					Namespace: repo.Namespace,
					Cluster:   s.kubeappsCluster,
				},
				Identifier: repo.Name,
				Plugin:     GetPluginDetail(),
			},
			Name: repo.Name,
			// TBD Flux HelmRepository CR doesn't have a designated field for description
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             repo.Spec.URL,
			Status:          repoStatus(repo),
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (s *Server) validateRepoUserManagedSecrets(
	ctx context.Context,
	repoName types.NamespacedName,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (string, error) {
	var secretRefTls, secretRefAuth string
	if tlsConfig != nil {
		if tlsConfig.GetCertAuthority() != "" {
			return "", status.Errorf(codes.InvalidArgument, "Secret Ref must be used with user managed secrets")
		} else if tlsConfig.GetSecretRef().GetName() != "" {
			secretRefTls = tlsConfig.GetSecretRef().GetName()
		}
	}

	if auth != nil {
		if auth.GetDockerCreds() != nil ||
			auth.GetHeader() != "" ||
			auth.GetTlsCertKey() != nil ||
			auth.GetUsernamePassword() != nil {
			return "", status.Errorf(codes.InvalidArgument, "Secret Ref must be used with user managed secrets")
		} else if auth.GetSecretRef().GetName() != "" {
			secretRefAuth = auth.GetSecretRef().GetName()
		}
	}

	var secretRef string
	if secretRefTls != "" && secretRefAuth != "" && secretRefTls != secretRefAuth {
		// flux repo spec only allows one secret per HelmRepository CRD
		return "", status.Errorf(
			codes.InvalidArgument, "TLS config secret and Auth secret must be the same")
	} else if secretRefTls != "" {
		secretRef = secretRefTls
	} else if secretRefAuth != "" {
		secretRef = secretRefAuth
	}

	if secretRef != "" {
		// check that the specified secret exists
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return "", err
		} else if _, err = typedClient.CoreV1().Secrets(repoName.Namespace).Get(ctx, secretRef, metav1.GetOptions{}); err != nil {
			return "", statuserror.FromK8sError("get", "secret", secretRef, err)
		}
		// TODO (gfichtenholt) also check that the data in the opaque secret corresponds
		// to specified auth type, e.g. if AuthType is
		// PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
		// check that the secret has "username" and "password" fields, etc.

		// TODO (gfichtenholt)
		// ref https://github.com/vmware-tanzu/kubeapps/pull/4353#discussion_r816332595
		// check whether flux supports typed secrets in addition to opaque secrets
		// https://kubernetes.io/docs/concepts/configuration/secret/#secret-types
		// If so, that cause certain validation to be done on the data (ie. ensuring that
		//	the "username" and "password" fields are present).
	}
	return secretRef, nil
}

func (s *Server) createRepoKubeappsManagedSecrets(
	ctx context.Context,
	repoName types.NamespacedName,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (string, error) {

	secret, err := newSecretFromTlsConfigAndAuth(repoName, tlsConfig, auth)
	if err != nil {
		return "", err
	}

	secretRef := ""
	if secret != nil {
		// create a secret first, if applicable
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return "", err
		} else if secret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return "", statuserror.FromK8sError("create", "secret", secret.GetName(), err)
		} else {
			secretRef = secret.GetName()
		}
	}
	return secretRef, nil
}

func (s *Server) updateRepoKubeappsManagedSecrets(
	ctx context.Context,
	repoName types.NamespacedName,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth,
	existingSecretRef *fluxmeta.LocalObjectReference) (string, error) {

	secret, err := newSecretFromTlsConfigAndAuth(repoName, tlsConfig, auth)
	if err != nil {
		return "", err
	}

	secretRef := ""
	typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
	if err != nil {
		return "", err
	}
	secretInterface := typedClient.CoreV1().Secrets(repoName.Namespace)
	if secret != nil {
		if existingSecretRef == nil {
			// create a secret first
			newSecret, err := secretInterface.Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return "", statuserror.FromK8sError("create", "secret", secret.GetGenerateName(), err)
			}
			secretRef = newSecret.GetName()
		} else {
			// TODO (gfichtenholt) we should optimize this to somehow tell if the existing secret
			// is the same (data-wise) as the new one and if so skip all this
			if err = secretInterface.Delete(ctx, existingSecretRef.Name, metav1.DeleteOptions{}); err != nil {
				return "", statuserror.FromK8sError("get", "secret", existingSecretRef.Name, err)
			}
			// create a new one
			newSecret, err := secretInterface.Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return "", statuserror.FromK8sError("update", "secret", secret.GetGenerateName(), err)
			}
			secretRef = newSecret.GetName()
		}
	} else if existingSecretRef != nil {
		if err = secretInterface.Delete(ctx, existingSecretRef.Name, metav1.DeleteOptions{}); err != nil {
			log.Errorf("Error deleting existing secret: [%s] due to %v", err)
		}
	}
	return secretRef, nil
}

func (s *Server) updateRepo(ctx context.Context, repoRef *corev1.PackageRepositoryReference, url string, interval uint32, tlsConfig *corev1.PackageRepositoryTlsConfig, auth *corev1.PackageRepositoryAuth) (*corev1.PackageRepositoryReference, error) {
	key := types.NamespacedName{Namespace: repoRef.GetContext().GetNamespace(), Name: repoRef.GetIdentifier()}
	repo, err := s.getRepoInCluster(ctx, key)
	if err != nil {
		return nil, err
	}

	// As Michael and I agreed 4/12/2022, initially we'll disallow updates to repos in
	// pending state to simplify the initial case, though we may implement support later.
	// Updates to non-pending repos (i.e. success or failed status) are allowed
	complete, _, _ := isHelmRepositoryReady(*repo)
	if !complete {
		return nil, status.Errorf(codes.Internal, "updates to repositories pending reconciliation are not supported")
	}

	if url == "" {
		return nil, status.Errorf(codes.InvalidArgument, "repository url may not be empty")
	}
	repo.Spec.URL = url

	// flux does not grok description yet

	if interval > 0 {
		repo.Spec.Interval = metav1.Duration{Duration: time.Duration(interval) * time.Second}
	} else {
		// interval is a required field
		repo.Spec.Interval = defaultPollInterval
	}

	if tlsConfig != nil && tlsConfig.InsecureSkipVerify {
		return nil, status.Errorf(codes.InvalidArgument, "TLS flag insecureSkipVerify is not supported")
	}

	var secretRef string
	if s.pluginConfig.UserManagedSecrets {
		if secretRef, err = s.validateRepoUserManagedSecrets(ctx, key, tlsConfig, auth); err != nil {
			return nil, err
		}
	} else {
		if secretRef, err = s.updateRepoKubeappsManagedSecrets(ctx, key, tlsConfig, auth, repo.Spec.SecretRef); err != nil {
			return nil, err
		}
	}

	if secretRef != "" {
		repo.Spec.SecretRef = &fluxmeta.LocalObjectReference{Name: secretRef}
	} else {
		repo.Spec.SecretRef = nil
	}

	repo.Spec.PassCredentials = auth != nil && auth.PassCredentials

	// get rid of the status field, since now there will be a new reconciliation
	// process and the current status no longer applies. metadata and spec I want
	// to keep, as they may have had added labels and/or annotations and/or
	// even other changes made by the user.
	repo.Status = sourcev1.HelmRepositoryStatus{}

	client, err := s.getClient(ctx, key.Namespace)
	if err != nil {
		return nil, err
	}
	if err = client.Update(ctx, repo); err != nil {
		return nil, statuserror.FromK8sError("update", "HelmRepository", key.String(), err)
	}

	log.V(4).Infof("Updated repository: %s", common.PrettyPrint(repo))

	return &corev1.PackageRepositoryReference{
		Context: &corev1.Context{
			Namespace: key.Namespace,
			Cluster:   s.kubeappsCluster,
		},
		Identifier: key.Name,
		Plugin:     GetPluginDetail(),
	}, nil
}

func (s *Server) deleteRepo(ctx context.Context, repoRef *corev1.PackageRepositoryReference) error {
	client, err := s.getClient(ctx, repoRef.Context.Namespace)
	if err != nil {
		return err
	}

	log.V(4).Infof("Deleting repo: [%s]", repoRef.Identifier)

	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repoRef.Identifier,
			Namespace: repoRef.Context.Namespace,
		},
	}

	// TODO (gfichtenholt) secrets, if applicable for kubeapps-managed secrets environment

	if err = client.Delete(ctx, repo); err != nil {
		return statuserror.FromK8sError("delete", "HelmRepository", repoRef.Identifier, err)
	}
	return nil
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
			// ref: https://github.com/vmware-tanzu/kubeapps/pull/3899#issuecomment-990446931
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
		URL:       repo.Spec.URL,
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

// this is only until https://github.com/vmware-tanzu/kubeapps/issues/3496
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
	if !checkRepoGeneration(repo) {
		return false
	}

	completed, success, _ := isHelmRepositoryReady(repo)
	return completed && success
}

// returns 3 things:
// - complete: whether the operation was completed
// - success: (only applicable when complete == true) whether the operation was successful or failed
// - reason: if present
// docs:
// 1. https://fluxcd.io/docs/components/source/helmrepositories/#status-examples
func isHelmRepositoryReady(repo sourcev1.HelmRepository) (complete bool, success bool, reason string) {
	// flux source-controller v1beta2 API made a change so that we can no longer
	// rely on a simple "metadata.generation" vs "status.observedGeneration" check for a
	// quick answer. The resource may now exist with "observedGeneration": -1 either in
	// pending or in a failed state. We need to distinguish between the two. Personally,
	// feels like a mistake to me.
	readyCond := meta.FindStatusCondition(repo.GetConditions(), fluxmeta.ReadyCondition)
	if readyCond != nil {
		if readyCond.Reason != "" {
			// this could be something like "reason": "Succeeded" i.e. not super-useful
			reason = readyCond.Reason
		}
		if readyCond.Message != "" {
			// whereas this could be something like: "message": 'invalid chart URL format'
			// i.e. a little more useful, so we'll just return them both
			reason += ": " + readyCond.Message
		}
		switch readyCond.Status {
		case metav1.ConditionTrue:
			return checkRepoGeneration(repo), true, reason
		case metav1.ConditionFalse:
			return true, false, reason
			// metav1.ConditionUnknown falls through
		}
	}
	return false, false, reason
}

func repoStatus(repo sourcev1.HelmRepository) *corev1.PackageRepositoryStatus {
	complete, success, reason := isHelmRepositoryReady(repo)
	s := &corev1.PackageRepositoryStatus{
		Ready:      complete && success,
		Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_UNSPECIFIED,
		UserReason: reason,
	}
	if !complete {
		s.Reason = corev1.PackageRepositoryStatus_STATUS_REASON_PENDING
	} else if success {
		s.Reason = corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS
	} else {
		s.Reason = corev1.PackageRepositoryStatus_STATUS_REASON_FAILED
	}
	return s
}

func checkRepoGeneration(repo sourcev1.HelmRepository) bool {
	generation := repo.GetGeneration()
	observedGeneration := repo.Status.ObservedGeneration
	return generation > 0 && generation == observedGeneration
}

// ref https://fluxcd.io/docs/components/source/helmrepositories/
func newFluxHelmRepo(
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

// this func is only used with kubeapps-managed secrets
func newSecretFromTlsConfigAndAuth(repoName types.NamespacedName,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*apiv1.Secret, error) {
	var secret *apiv1.Secret
	if tlsConfig != nil {
		if tlsConfig.GetSecretRef() != nil {
			return nil, status.Errorf(codes.InvalidArgument, "SecretRef may not be used with kubeapps managed secrets")
		}
		caCert := tlsConfig.GetCertAuthority()
		if caCert != "" {
			secret = common.NewLocalOpaqueSecret(repoName.Name + "-")
			secret.Data["caFile"] = []byte(caCert)
		}
	}
	if auth != nil {
		if auth.GetSecretRef() != nil {
			return nil, status.Errorf(codes.InvalidArgument, "SecretRef may not be used with kubeapps managed secrets")
		}
		if secret == nil {
			secret = common.NewLocalOpaqueSecret(repoName.Name + "-")
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
				secret.Data[".dockerconfigjson"] = common.DockerCredentialsToSecretData(dc)
			} else {
				return nil, status.Errorf(codes.Internal, "Docker credentials configuration is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED:
			return nil, nil
		default:
			return nil, status.Errorf(codes.Internal, "Unexpected package repository authentication type: %q", auth.Type)
		}
	}
	return secret, nil
}

func getRepoTlsConfigAndAuthWithUserManagedSecrets(secret *apiv1.Secret) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
	var tlsConfig *corev1.PackageRepositoryTlsConfig
	auth := &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
	}

	if _, ok := secret.Data["caFile"]; ok {
		tlsConfig = &corev1.PackageRepositoryTlsConfig{
			// flux plug in doesn't support this option
			InsecureSkipVerify: false,
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: secret.Name,
					Key:  "caFile",
				},
			},
		}
	}
	if _, ok := secret.Data["certFile"]; ok {
		if _, ok = secret.Data["keyFile"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{Name: secret.Name},
			}
		}
	} else if _, ok := secret.Data["username"]; ok {
		if _, ok = secret.Data["password"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{Name: secret.Name},
			}
		}
	} else if _, ok := secret.Data[".dockerconfigjson"]; ok {
		auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
		auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
			SecretRef: &corev1.SecretKeyReference{Name: secret.Name},
		}
	} else {
		log.Warning("Unrecognized type of secret [%s]", secret.Name)
	}
	return tlsConfig, auth, nil
}

func getRepoTlsConfigAndAuthWithKubeappsManagedSecrets(secret *apiv1.Secret) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
	var tlsConfig *corev1.PackageRepositoryTlsConfig
	auth := &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
	}

	if caFile, ok := secret.Data["caFile"]; ok {
		tlsConfig = &corev1.PackageRepositoryTlsConfig{
			// flux plug in doesn't support this option
			InsecureSkipVerify: false,
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
				CertAuthority: string(caFile),
			},
		}
	}

	if certFile, ok := secret.Data["certFile"]; ok {
		if keyFile, ok := secret.Data["keyFile"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_TlsCertKey{
				TlsCertKey: &corev1.TlsCertKey{
					Cert: string(certFile),
					Key:  string(keyFile),
				},
			}
		}
	} else if username, ok := secret.Data["username"]; ok {
		if pwd, ok := secret.Data["password"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_UsernamePassword{
				UsernamePassword: &corev1.UsernamePassword{
					Username: string(username),
					Password: string(pwd),
				},
			}
		}
	} else if configStr, ok := secret.Data[".dockerconfigjson"]; ok {
		dc, err := common.SecretDataToDockerCredentials(string(configStr))
		if err != nil {
			return nil, nil, err
		}
		auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
		auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_DockerCreds{
			DockerCreds: dc,
		}
	} else {
		log.Warning("Unrecognized type of secret [%s]", secret.Name)
	}
	return tlsConfig, auth, nil
}
