// Temporary file for debugging exactly what the exact HTTP requests to
// and responses from a remote OCI registry look like.
// Kinf of like a network sniffer. This file will eventually go away

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"helm.sh/helm/v3/pkg/registry"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	log "k8s.io/klog/v2"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/helmpath"
	dockerauth "oras.land/oras-go/pkg/auth/docker"
	orascontext "oras.land/oras-go/pkg/context"
	orasregistry "oras.land/oras-go/pkg/registry"
	registryremote "oras.land/oras-go/pkg/registry/remote"
	registryauth "oras.land/oras-go/pkg/registry/remote/auth"
)

// Retrieve list of repository tags
// impl ref https://github.com/helm/helm/blob/657850e44b880cca43d0606ebf5a54eb75362c3f/pkg/registry/client.go#L579
func debugTagList(ref string) error {
	log.Infof("+debugTagList(%s)", ref)

	parsedReference, err := orasregistry.ParseReference(ref)
	if err != nil {
		return err
	}
	// given ref like this
	//   ghcr.io/stefanprodan/charts/podinfo
	// will return
	//  {
	//    "Registry": "ghcr.io",
	//    "Repository": "stefanprodan/charts/podinfo",
	//    "Reference": ""
	// }
	log.Infof("debugTags parsedReference(%s): %s", ref, common.PrettyPrint(parsedReference))

	credentialsFile := helmpath.ConfigPath(registry.CredentialsFileBasename)
	log.Infof("debugTags credentials file: %s", credentialsFile)

	authClient, err := dockerauth.NewClientWithDockerFallback(credentialsFile)
	if err != nil {
		return err
	}

	registryAuthorizer := &registryauth.Client{
		Header: http.Header{"User-Agent": {"Helm/3.9.0"}},
		Cache:  registryauth.DefaultCache,
		Credential: func(ctx context.Context, reg string) (registryauth.Credential, error) {
			dockerClient, ok := authClient.(*dockerauth.Client)
			if !ok {
				return registryauth.EmptyCredential, errors.New("unable to obtain docker client")
			}

			username, password, err := dockerClient.Credential(reg)
			if err != nil {
				return registryauth.EmptyCredential, errors.New("unable to retrieve credentials")
			}

			printPwd := password
			if len(printPwd) > 3 {
				printPwd = printPwd[0:3] + "..."
			}
			log.Infof("=======> debugTags: registryAuthorizer: username: [%s] password: [%s]",
				username, printPwd)

			// A blank returned username and password value is a bearer token
			if username == "" && password != "" {
				return registryauth.Credential{
					RefreshToken: password,
				}, nil
			}
			return registryauth.Credential{
				Username: username,
				Password: password,
			}, nil
		},
	}

	repository := registryremote.Repository{
		Reference: parsedReference,
		Client:    registryAuthorizer,
	}

	ctxFn := func(out io.Writer, debug bool) context.Context {
		if !debug {
			return orascontext.Background()
		}
		ctx := orascontext.WithLoggerFromWriter(context.Background(), out)
		orascontext.GetLogger(ctx).Logger.SetLevel(logrus.DebugLevel)
		return ctx
	}

	// ref https://github.com/oras-project/oras-go/blob/main/registry/remote/url.go
	// buildScheme returns HTTP scheme used to access the remote registry.
	buildScheme := func(plainHTTP bool) string {
		if plainHTTP {
			return "http"
		}
		return "https"
	}

	// buildRepositoryBaseURL builds the base endpoint of the remote repository.
	// Format: <scheme>://<registry>/v2/<repository>
	//buildRepositoryBaseURL :=
	_ = func(plainHTTP bool, ref orasregistry.Reference) string {
		return fmt.Sprintf("%s://%s/v2/%s", buildScheme(plainHTTP), ref.Host(), ref.Repository)
	}

	// buildRepositoryTagListURL builds the URL for accessing the tag list API.
	// Format: <scheme>://<registry>/v2/<repository>/tags/list
	// Reference: https://docs.docker.com/registry/spec/api/#tags
	buildRepositoryTagListURL := func(plainHTTP bool, ref orasregistry.Reference) string {
		//return buildRepositoryBaseURL(plainHTTP, ref) + "/tags/list"
		//return "https://ghcr.io/v2/_catalog"
		return "https://ghcr.io/v2/"
	}

	tags := func(ctx context.Context, url string) (string, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", err
		}
		if repository.TagListPageSize > 0 {
			q := req.URL.Query()
			q.Set("n", strconv.Itoa(repository.TagListPageSize))
			req.URL.RawQuery = q.Encode()
		}

		log.Infof("debugTags: +HTTP %s request:\nURL:\n%s\npretty print:\n%s",
			req.Method,
			req.URL.String(),
			common.PrettyPrint(req))
		resp, err := repository.Client.Do(req)
		if err != nil {
			log.Infof("debugTags: -HTTP GET response: raised err=%v", err)
			return "", err
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		log.Infof("debugTags: -HTTP %s response: code:%s\nbody:\n%s\nheaders:\n%s\nerr=%v",
			req.Method,
			resp.Status,
			respBody,
			resp.Header,
			err)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		return "", nil
	}

	url := buildRepositoryTagListURL(repository.PlainHTTP, repository.Reference)
	tags(ctxFn(io.Discard, true), url)

	return nil
}
