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
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// see docs at https://fluxcd.io/docs/components/source/ and
	// https://fluxcd.io/docs/components/helm/api/
	fluxHelmRepositories = "helmrepositories"
	redactedString       = "REDACTED"
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
	for r := range repoList {
		repo := repoList[r] // avoid implicit memory aliasing
		// first check if repo is in ready state
		if !isRepoReady(repo) {
			// just skip it
			continue
		}
		name, err := common.NamespacedName(&repo)
		if err != nil {
			// #nosec G104
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

	chartsUntyped, err := s.repoCache.GetMultiple(repoNames)
	if err != nil {
		return nil, err
	}

	chartsTyped := make(map[string][]models.Chart)
	for key, value := range chartsUntyped {
		if value == nil {
			chartsTyped[key] = nil
		} else {
			typedValue, err := s.repoCacheEntryFromUntyped(key, value)
			if err != nil {
				return nil, err
			} else if typedValue == nil {
				chartsTyped[key] = nil
			} else {
				chartsTyped[key] = typedValue.Charts
			}
		}
	}
	return chartsTyped, nil
}

func (s *Server) repoCacheEntryFromUntyped(key string, value interface{}) (*repoCacheEntryValue, error) {
	if value == nil {
		return nil, nil
	}
	typedValue, ok := value.(repoCacheEntryValue)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			"unexpected value fetched from cache: type: [%s], value: [%v]",
			reflect.TypeOf(value), value)
	}
	if typedValue.Type == "oci" {
		// ref https://github.com/vmware-tanzu/kubeapps/issues/5007#issuecomment-1217293240
		// helm OCI chart repos are not automatically updated when the
		// state on remote changes. So we will force new checksum
		// computation and update local cache if needed
		value, err := s.repoCache.ForceAndFetch(key)
		if err != nil {
			return nil, err
		} else if value != nil {
			typedValue, ok = value.(repoCacheEntryValue)
			if !ok {
				return nil, status.Errorf(
					codes.Internal,
					"unexpected value fetched from cache: type: [%s], value: [%v]",
					reflect.TypeOf(value), value)
			}
		}
	}
	return &typedValue, nil
}

func (s *Server) httpClientOptionsForRepo(ctx context.Context, repoName types.NamespacedName) (*common.HttpClientOptions, error) {
	repo, err := s.getRepoInCluster(ctx, repoName)
	if err != nil {
		return nil, err
	}
	sink := s.newRepoEventSink()
	return sink.clientOptionsForHttpRepo(ctx, *repo)
}

func (s *Server) newRepo(ctx context.Context, request *corev1.AddPackageRepositoryRequest) (*corev1.PackageRepositoryReference, error) {
	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request Name provided")
	}

	if request.GetNamespaceScoped() {
		return nil, status.Errorf(codes.Unimplemented, "namespaced-scoped repositories are not supported")
	}

	typ := request.GetType()
	if typ != "helm" && typ != sourcev1.HelmRepositoryTypeOCI {
		return nil, status.Errorf(codes.Unimplemented, "repository type [%s] not supported", typ)
	}

	url := request.GetUrl()
	tlsConfig := request.GetTlsConfig()
	if url == "" {
		return nil, status.Errorf(codes.InvalidArgument, "repository url may not be empty")
	} else if tlsConfig != nil && tlsConfig.InsecureSkipVerify {
		// ref https://github.com/fluxcd/source-controller/issues/807
		return nil, status.Errorf(codes.InvalidArgument, "TLS flag insecureSkipVerify is not supported")
	}

	name := types.NamespacedName{Name: request.Name, Namespace: request.Context.Namespace}
	auth := request.GetAuth()
	var secret *apiv1.Secret
	var err error
	if s.pluginConfig.UserManagedSecrets {
		if secret, err = s.validateUserManagedRepoSecret(ctx, name, typ, tlsConfig, auth); err != nil {
			return nil, err
		}
	} else {
		// a bit of catch 22: I need to create a secret first, so that I can create a repo that references it
		// but then I need to set the owner reference on this secret to the repo. In has to be done
		// in that order because to set an owner ref you need object (i.e. repo) UID, which you only get
		// once the object's been created
		if secret, err = s.createKubeappsManagedRepoSecret(ctx, name, typ, tlsConfig, auth); err != nil {
			return nil, err
		}
	}

	passCredentials := auth != nil && auth.PassCredentials
	interval := request.GetInterval()

	// Get Flux-specific values
	provider := ""
	var customDetail *v1alpha1.FluxPackageRepositoryCustomDetail
	if request.CustomDetail != nil {
		customDetail = &v1alpha1.FluxPackageRepositoryCustomDetail{}
		if err := request.CustomDetail.UnmarshalTo(customDetail); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "customDetail could not be parsed due to: %v", err)
		}
		log.Infof("fluxv2 customDetail: [%v]", customDetail)
		provider = customDetail.Provider
	}

	if fluxRepo, err := newFluxHelmRepo(name, typ, url, interval, secret, passCredentials, provider); err != nil {
		return nil, err
	} else if client, err := s.getClient(ctx, name.Namespace); err != nil {
		return nil, err
	} else if err = client.Create(ctx, fluxRepo); err != nil {
		return nil, statuserror.FromK8sError("create", "HelmRepository", name.String(), err)
	} else {
		if !s.pluginConfig.UserManagedSecrets {
			if err = s.setOwnerReferencesForRepoSecret(ctx, secret, fluxRepo); err != nil {
				return nil, err
			}
		}
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
	typ := repo.Spec.Type
	if typ == "" {
		typ = "helm"
	}

	// Get Fluxv2-specific values
	var customDetail *anypb.Any
	// For now: this is somewhat my subjective call to filter out "generic" (default) ones
	// because otherwise any repo created with an unset provider will come back from flux
	// as "generic" and therefore the PackageRepositoryDetail instance returned by this func
	// will have a FluxPackageRepositoryCustomDetail in it. Flux spec already clearly states
	// If you do not specify .spec.provider, it defaults to generic.
	// https://fluxcd.io/flux/components/source/helmrepositories/#provider
	if repo.Spec.Provider != "" && repo.Spec.Provider != sourcev1.GenericOCIProvider {
		if customDetail, err = anypb.New(&v1alpha1.FluxPackageRepositoryCustomDetail{
			Provider: repo.Spec.Provider,
		}); err != nil {
			return nil, status.Errorf(codes.Internal, "custom detail could not be marshalled due to: %v", err)
		}
	}

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
		// TODO (gfichtenholt) Flux HelmRepository CR doesn't have a designated field for description
		Description:     "",
		NamespaceScoped: false,
		Type:            typ,
		Url:             repo.Spec.URL,
		Interval:        pkgutils.FromDuration(&repo.Spec.Interval),
		TlsConfig:       tlsConfig,
		Auth:            auth,
		Status:          repoStatus(*repo),
		CustomDetail:    customDetail,
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
		typ := repo.Spec.Type
		if typ == "" {
			typ = "helm"
		}

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
			// TODO (gfichtenholt) Flux HelmRepository CR doesn't have a designated field for description
			Description:     "",
			NamespaceScoped: false,
			Type:            typ,
			Url:             repo.Spec.URL,
			Status:          repoStatus(repo),
			RequiresAuth:    repo.Spec.SecretRef != nil,
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (s *Server) validateUserManagedRepoSecret(
	ctx context.Context,
	repoName types.NamespacedName,
	repoType string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*apiv1.Secret, error) {
	var secretRefTls, secretRefAuth string
	if tlsConfig != nil {
		if tlsConfig.GetCertAuthority() != "" {
			return nil, status.Errorf(codes.InvalidArgument, "Secret Ref must be used with user managed secrets")
		} else if tlsConfig.GetSecretRef().GetName() != "" {
			secretRefTls = tlsConfig.GetSecretRef().GetName()
		}
	}

	if auth != nil {
		if auth.GetDockerCreds() != nil ||
			auth.GetHeader() != "" ||
			auth.GetTlsCertKey() != nil ||
			auth.GetUsernamePassword() != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Secret Ref must be used with user managed secrets")
		} else if auth.GetSecretRef().GetName() != "" {
			secretRefAuth = auth.GetSecretRef().GetName()
		}
	}

	var secretRef string
	if secretRefTls != "" && secretRefAuth != "" && secretRefTls != secretRefAuth {
		// flux repo spec only allows one secret per HelmRepository CRD
		return nil, status.Errorf(
			codes.InvalidArgument, "TLS config secret and Auth secret must be the same")
	} else if secretRefTls != "" {
		secretRef = secretRefTls
	} else if secretRefAuth != "" {
		secretRef = secretRefAuth
	}

	var secret *apiv1.Secret
	if secretRef != "" {
		// check that the specified secret exists
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return nil, err
		} else if secret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Get(ctx, secretRef, metav1.GetOptions{}); err != nil {
			return nil, statuserror.FromK8sError("get", "secret", secretRef, err)
		} else {
			// also check that the data in the opaque secret corresponds
			// to specified auth type, e.g. if AuthType is
			// PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			// check that the secret has "username" and "password" fields, etc.
			// it appears flux does not care about the k8s secret type (opaque vs tls vs basic-auth, etc.)
			// https://github.com/fluxcd/source-controller/blob/bc5a47e821562b1c4f9731acd929b8d9bd23b3a8/controllers/helmrepository_controller.go#L357
			if secretRefTls != "" && secret.Data["caFile"] == nil {
				return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing field 'caFile'", secretRef)
			}
			if secretRefAuth != "" {
				switch auth.Type {
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
					if secret.Data["username"] == nil || secret.Data["password"] == nil {
						return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing fields 'username' and/or 'password'", secretRef)
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
					if secret.Data["keyFile"] == nil || secret.Data["certFile"] == nil {
						return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing fields 'keyFile' and/or 'certFile'", secretRef)
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
					if repoType == sourcev1.HelmRepositoryTypeOCI {
						if secret.Data[apiv1.DockerConfigJsonKey] == nil {
							return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing field '%s'", secretRef, apiv1.DockerConfigJsonKey)
						}
					} else {
						return nil, status.Errorf(codes.Internal, "Package repository authentication type %q is not supported", auth.Type)
					}
				default:
					return nil, status.Errorf(codes.Internal, "Package repository authentication type %q is not supported", auth.Type)
				}
			}
		}

		// ref https://github.com/vmware-tanzu/kubeapps/pull/4353#discussion_r816332595
		// check whether flux supports typed secrets in addition to opaque secrets
		// https://kubernetes.io/docs/concepts/configuration/secret/#secret-types
		// update: flux currently does not care about secret type, just what is in the data map.
	}
	return secret, nil
}

func (s *Server) createKubeappsManagedRepoSecret(
	ctx context.Context,
	repoName types.NamespacedName,
	typ string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*apiv1.Secret, error) {

	secret, _, err := newSecretFromTlsConfigAndAuth(repoName, typ, tlsConfig, auth)
	if err != nil {
		return nil, err
	}

	if secret != nil {
		// create a secret first, if applicable
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return nil, err
		} else if secret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return nil, statuserror.FromK8sError("create", "secret", secret.GetName(), err)
		}
	}
	return secret, nil
}

// using owner references on the secret so that it can be
// (1) cleaned up automatically and/or
// (2) enable some control (ie. if I add a secret manually
//   via kubectl before running kubeapps, it won't get deleted just
// because Kubeapps is deleting it)?
// see https://github.com/vmware-tanzu/kubeapps/pull/4630#discussion_r861446394 for details
func (s *Server) setOwnerReferencesForRepoSecret(
	ctx context.Context,
	secret *apiv1.Secret,
	repo *sourcev1.HelmRepository) error {

	if repo.Spec.SecretRef != nil && secret != nil {
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return err
		} else {
			secretsInterface := typedClient.CoreV1().Secrets(repo.Namespace)
			secret.OwnerReferences = []metav1.OwnerReference{
				*metav1.NewControllerRef(
					repo,
					schema.GroupVersionKind{
						Group:   sourcev1.GroupVersion.Group,
						Version: sourcev1.GroupVersion.Version,
						Kind:    sourcev1.HelmRepositoryKind,
					}),
			}
			if _, err := secretsInterface.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
				return statuserror.FromK8sError("update", "secrets", secret.Name, err)
			}
		}
	}
	return nil
}

// returns 3 things:
// secret     - either an existing or newly created secret, or nil if there should
//              be no secret associated with a repository
// updateRepo - boolean indicating whether or not a HelmRepository CRD needs to be updated
// err        - when bad things happen

func (s *Server) updateKubeappsManagedRepoSecret(
	ctx context.Context,
	repoName types.NamespacedName,
	typ string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth,
	existingSecretRef *fluxmeta.LocalObjectReference) (secret *apiv1.Secret, updateRepo bool, err error) {

	secret, isSameSecret, err := newSecretFromTlsConfigAndAuth(repoName, typ, tlsConfig, auth)
	if err != nil {
		return nil, false, err
	} else if isSameSecret {
		return nil, false, nil
	}

	typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
	if err != nil {
		return nil, false, err
	}
	secretInterface := typedClient.CoreV1().Secrets(repoName.Namespace)
	if secret != nil {
		if existingSecretRef == nil {
			// create a secret first
			newSecret, err := secretInterface.Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return nil, false, statuserror.FromK8sError("create", "secret", secret.GetGenerateName(), err)
			}
			return newSecret, true, nil
		} else {
			// TODO (gfichtenholt) we should optimize this to somehow tell if the existing secret
			// is the same (data-wise) as the new one and if so skip all this
			if err = secretInterface.Delete(ctx, existingSecretRef.Name, metav1.DeleteOptions{}); err != nil {
				return nil, false, statuserror.FromK8sError("delete", "secret", existingSecretRef.Name, err)
			}
			// create a new one
			newSecret, err := secretInterface.Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return nil, false, statuserror.FromK8sError("create", "secret", secret.GetGenerateName(), err)
			}
			return newSecret, true, nil
		}
	} else if existingSecretRef != nil {
		if err = secretInterface.Delete(ctx, existingSecretRef.Name, metav1.DeleteOptions{}); err != nil {
			log.Errorf("Error deleting existing secret: [%s] due to %v", err)
		}
	}
	return secret, true, nil
}

func (s *Server) updateRepo(ctx context.Context, repoRef *corev1.PackageRepositoryReference, url string, interval string, tlsConfig *corev1.PackageRepositoryTlsConfig, auth *corev1.PackageRepositoryAuth) (*corev1.PackageRepositoryReference, error) {
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

	// flux does not grok repository description yet
	// the only field in customdetail is "provider" and I don't see the need to
	// have the user update that. Its not like one repository is going to move from
	// GCP to AWS.

	if interval != "" {
		if duration, err := pkgutils.ToDuration(interval); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "interval is invalid: %v", err)
		} else {
			repo.Spec.Interval = *duration
		}
	} else {
		// interval is a required field
		repo.Spec.Interval = defaultPollInterval
	}

	if tlsConfig != nil && tlsConfig.InsecureSkipVerify {
		// ref https://github.com/fluxcd/source-controller/issues/807
		return nil, status.Errorf(codes.InvalidArgument, "TLS flag insecureSkipVerify is not supported")
	}

	var secret *apiv1.Secret
	var updateRepoSecret bool
	if s.pluginConfig.UserManagedSecrets {
		if secret, err = s.validateUserManagedRepoSecret(ctx, key, repo.Spec.Type, tlsConfig, auth); err != nil {
			return nil, err
		}
	} else {
		if secret, updateRepoSecret, err = s.updateKubeappsManagedRepoSecret(
			ctx, key, repo.Spec.Type, tlsConfig, auth, repo.Spec.SecretRef); err != nil {
			return nil, err
		}
	}

	if s.pluginConfig.UserManagedSecrets || updateRepoSecret {
		if secret != nil {
			repo.Spec.SecretRef = &fluxmeta.LocalObjectReference{Name: secret.Name}
		} else {
			repo.Spec.SecretRef = nil
		}
	}

	repo.Spec.PassCredentials = auth != nil && auth.PassCredentials

	// get rid of the status field, since now there will be a new reconciliation
	// process and the current status no longer applies. metadata and spec I want
	// to keep, as they may have had added labels and/or annotations and/or
	// even other changes made by the user.
	repo.Status = sourcev1.HelmRepositoryStatus{}

	if client, err := s.getClient(ctx, key.Namespace); err != nil {
		return nil, err
	} else if err = client.Update(ctx, repo); err != nil {
		return nil, statuserror.FromK8sError("update", "HelmRepository", key.String(), err)
	} else if updateRepoSecret && secret != nil {
		// new secret => will need to set the owner
		if err = s.setOwnerReferencesForRepoSecret(ctx, secret, repo); err != nil {
			return nil, err
		}
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

	// For kubeapps-managed secrets environment secrets will be deleted (garbage-collected)
	// when the owner repo is deleted

	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repoRef.Identifier,
			Namespace: repoRef.Context.Namespace,
		},
	}
	if err = client.Delete(ctx, repo); err != nil {
		return statuserror.FromK8sError("delete", "HelmRepository", repoRef.Identifier, err)
	} else {
		return nil
	}
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
	Type     string // if missing, repo is assumed to be regular old HTTP
	Charts   []models.Chart
}

// onAddRepo essentially tells the cache whether to and what to store for a given key
func (s *repoEventSink) onAddRepo(key string, obj ctrlclient.Object) (interface{}, bool, error) {
	log.V(4).Info("+onAddRepo(%s)", key)
	defer log.V(4).Info("-onAddRepo()")

	if repo, ok := obj.(*sourcev1.HelmRepository); !ok {
		return nil, false, fmt.Errorf("expected an instance of *sourcev1.HelmRepository, got: %s", reflect.TypeOf(obj))
	} else if isRepoReady(*repo) {
		if repo.Spec.Type == sourcev1.HelmRepositoryTypeOCI {
			return s.onAddOciRepo(*repo)
		} else {
			return s.onAddHttpRepo(*repo)
		}
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
func (s *repoEventSink) onAddHttpRepo(repo sourcev1.HelmRepository) ([]byte, bool, error) {
	if artifact := repo.GetArtifact(); artifact != nil {
		if checksum := artifact.Checksum; checksum == "" {
			return nil, false, status.Errorf(codes.Internal,
				"expected field status.artifact.checksum not found on HelmRepository\n[%s]",
				common.PrettyPrint(repo))
		} else {
			return s.indexAndEncode(checksum, repo)
		}
	} else {
		return nil, false, status.Errorf(codes.Internal,
			"expected field status.artifact not found on HelmRepository\n[%s]",
			common.PrettyPrint(repo))
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
		if opts, err := s.clientOptionsForHttpRepo(context.Background(), repo); err != nil {
			// ref: https://github.com/vmware-tanzu/kubeapps/pull/3899#issuecomment-990446931
			// I don't want this func to fail onAdd/onModify() if we can't read
			// the corresponding secret due to something like default RBAC settings:
			// "secrets "podinfo-basic-auth-secret" is forbidden:
			// User "system:serviceaccount:kubeapps:kubeapps-internal-kubeappsapis" cannot get
			// resource "secrets" in API group "" in the namespace "default"
			// So we still finish the indexing of the repo but skip the charts
			log.Errorf("Failed to read secret for repo due to: %+v", err)
		} else {
			fn := downloadHttpChartFn(opts)
			if err = s.chartCache.SyncCharts(charts, fn); err != nil {
				return nil, false, err
			}
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
	// shallow = false => 12-13 sec, so deep copy adds 50% to cost, but we need it
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
	// note that we are returning an array of model.Chart, each of which has an
	// array of model.ChartVersions, which in turn, only has those fields initialized that
	// can be read from index.yaml. Fields like readme, schema, values are empty at this point.
	// They do not get stored in the repo cache. They get stored in the chart cache
	// in a .tgz file
	return charts, nil
}

// onModifyRepo essentially tells the cache whether or not to and what to store for a given key
func (s *repoEventSink) onModifyRepo(key string, obj ctrlclient.Object, oldValue interface{}) (interface{}, bool, error) {
	if repo, ok := obj.(*sourcev1.HelmRepository); !ok {
		return nil, false, fmt.Errorf("expected an instance of *sourcev1.HelmRepository, got: %s", reflect.TypeOf(obj))
	} else if isRepoReady(*repo) {
		// first check the repo is ready

		if repo.Spec.Type == sourcev1.HelmRepositoryTypeOCI {
			return s.onModifyOciRepo(key, oldValue, *repo)
		} else {
			return s.onModifyHttpRepo(key, oldValue, *repo)
		}
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.V(4).Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

func (s *repoEventSink) onModifyHttpRepo(key string, oldValue interface{}, repo sourcev1.HelmRepository) ([]byte, bool, error) {
	// We should to compare checksums on what's stored in the cache
	// vs the modified object to see if the contents has really changed before embarking on
	// expensive operation indexOneRepo() below.
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
		return s.indexAndEncode(newChecksum, repo)
	} else {
		// skip because the content did not change
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

func (s *repoEventSink) getRepoSecret(ctx context.Context, repo sourcev1.HelmRepository) (*apiv1.Secret, error) {
	if repo.Spec.SecretRef == nil {
		return nil, nil
	}
	secretName := repo.Spec.SecretRef.Name
	if secretName == "" {
		return nil, nil
	}
	if s == nil || s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "unexpected state in clientGetter instance")
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
	return secret, err
}

// The reason I do this here is to set up auth that may be needed to fetch chart tarballs by
// ChartCache
func (s *repoEventSink) clientOptionsForHttpRepo(ctx context.Context, repo sourcev1.HelmRepository) (*common.HttpClientOptions, error) {
	if secret, err := s.getRepoSecret(ctx, repo); err == nil && secret != nil {
		return common.HttpClientOptionsFromSecret(*secret)
	} else {
		return nil, err
	}
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
	typ string,
	url string,
	interval string,
	secret *apiv1.Secret,
	passCredentials bool,
	provider string) (*sourcev1.HelmRepository, error) {
	pollInterval := defaultPollInterval
	if interval != "" {
		if duration, err := pkgutils.ToDuration(interval); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "interval is invalid: %v", err)
		} else {
			pollInterval = *duration
		}
	}
	fluxRepo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName.Name,
			Namespace: targetName.Namespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:      url,
			Interval: pollInterval,
		},
	}
	if typ == sourcev1.HelmRepositoryTypeOCI {
		fluxRepo.Spec.Type = sourcev1.HelmRepositoryTypeOCI
	}
	if secret != nil {
		fluxRepo.Spec.SecretRef = &fluxmeta.LocalObjectReference{
			Name: secret.Name,
		}
	}
	if passCredentials {
		fluxRepo.Spec.PassCredentials = true
	}
	if provider != "" {
		fluxRepo.Spec.Provider = provider
	}
	return fluxRepo, nil
}

// this func is only used with kubeapps-managed secrets
func newSecretFromTlsConfigAndAuth(repoName types.NamespacedName,
	typ string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (secret *apiv1.Secret, isSameSecret bool, err error) {
	if tlsConfig != nil {
		if tlsConfig.GetSecretRef() != nil {
			return nil, false, status.Errorf(codes.InvalidArgument, "SecretRef may not be used with kubeapps managed secrets")
		}
		caCert := tlsConfig.GetCertAuthority()
		if caCert == redactedString {
			isSameSecret = true
		} else if caCert != "" {
			secret = common.NewLocalOpaqueSecret(repoName)
			secret.Data["caFile"] = []byte(caCert)
		}
	}
	if auth != nil {
		if auth.GetSecretRef() != nil {
			return nil, false, status.Errorf(codes.InvalidArgument, "SecretRef may not be used with kubeapps managed secrets")
		}
		if secret == nil {
			if auth.Type == corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON {
				secret = common.NewLocalDockerConfigJsonSecret(repoName)
			} else {
				secret = common.NewLocalOpaqueSecret(repoName)
			}
		}
		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			if unp := auth.GetUsernamePassword(); unp != nil {
				if unp.Username == redactedString && unp.Password == redactedString {
					isSameSecret = true
				} else {
					secret.Data["username"] = []byte(unp.Username)
					secret.Data["password"] = []byte(unp.Password)
				}
			} else {
				return nil, false, status.Errorf(codes.Internal, "Username/Password configuration is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
			if ck := auth.GetTlsCertKey(); ck != nil {
				if ck.Cert == redactedString && ck.Key == redactedString {
					isSameSecret = true
				} else {
					secret.Data["certFile"] = []byte(ck.Cert)
					secret.Data["keyFile"] = []byte(ck.Key)
				}
			} else {
				return nil, false, status.Errorf(codes.Internal, "TLS Cert/Key configuration is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if typ == sourcev1.HelmRepositoryTypeOCI {
				if dc := auth.GetDockerCreds(); dc != nil {
					if dc.Username == redactedString && dc.Password == redactedString && dc.Server == redactedString {
						isSameSecret = true
					} else {
						secret.Data = map[string][]byte{
							apiv1.DockerConfigJsonKey: []byte(`{"auths":{"` +
								dc.Server + `":{"` +
								`auth":"` + base64.StdEncoding.EncodeToString([]byte(dc.Username+":"+dc.Password)) + `"}}}`),
						}
					}
				} else {
					return nil, false, status.Errorf(codes.Internal, "Docker credentials configuration is missing")
				}
			} else {
				return nil, false, status.Errorf(codes.Internal, "Unsupported package repository authentication type: %q", auth.Type)
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
			corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
			return nil, false, status.Errorf(codes.Internal, "Package repository authentication type %q is not supported", auth.Type)
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED:
			return nil, true, nil
		default:
			return nil, false, status.Errorf(codes.Internal, "Unsupported package repository authentication type: %q", auth.Type)
		}
	}
	return secret, isSameSecret, nil
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
	} else {
		log.Warning("Unrecognized type of secret [%s]", secret.Name)
	}
	return tlsConfig, auth, nil
}

// TODO (gfichtenolt) Per slack discussion
// In fact, keeping the existing API might mean we could return exactly what it already does today
//(i.e. all secrets) if called with an extra explicit option (includeSecrets=true in the request
// message, not sure, similar to kubectl  config view --raw) and by default the secrets are REDACTED
// as you mention? This would mean clients will by default see only REDACTED secrets,
// but can request the full sensitive data when necessary?
func getRepoTlsConfigAndAuthWithKubeappsManagedSecrets(secret *apiv1.Secret) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
	var tlsConfig *corev1.PackageRepositoryTlsConfig
	auth := &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
	}

	if _, ok := secret.Data["caFile"]; ok {
		tlsConfig = &corev1.PackageRepositoryTlsConfig{
			// flux plug in doesn't support InsecureSkipVerify option
			InsecureSkipVerify: false,
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
				CertAuthority: redactedString,
			},
		}
	}

	if _, ok := secret.Data["certFile"]; ok {
		if _, ok := secret.Data["keyFile"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_TlsCertKey{
				TlsCertKey: &corev1.TlsCertKey{
					Cert: redactedString,
					Key:  redactedString,
				},
			}
		}
	} else if _, ok := secret.Data["username"]; ok {
		if _, ok := secret.Data["password"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_UsernamePassword{
				UsernamePassword: &corev1.UsernamePassword{
					Username: redactedString,
					Password: redactedString,
				},
			}
		}
	} else if _, ok := secret.Data[apiv1.DockerConfigJsonKey]; ok {
		auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
		auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_DockerCreds{
			DockerCreds: &corev1.DockerCredentials{
				Username: redactedString,
				Password: redactedString,
				Server:   redactedString,
			},
		}
	} else {
		log.Warning("Unrecognized type of secret: [%s]", secret.Name)
	}
	return tlsConfig, auth, nil
}
