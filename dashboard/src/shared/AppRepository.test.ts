// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import * as moxios from "moxios";
import { AppRepository } from "./AppRepository";
import { axiosWithAuth } from "./AxiosInstance";
import * as url from "./url";
import { IPkgRepoFormData, IPkgRepositoryFilter } from "shared/types";
import { PackageRepositoryAuth_PackageRepositoryAuthType } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";

describe("AppRepository", () => {
  const cluster = "cluster";
  const namespace = "namespace";
  const pkgRepoFormData = {
    name: "repo-test",
    url: "repo-url",
    type: "repo-type",
    description: "repo-description",
    authHeader: "repo-authHeader",
    customCA: "repo-customCA",
    customDetails: {
      dockerRegistrySecrets: ["repo-secret1"],
      ociRepositories: ["oci-repo1"],
      filterRule: {
        jq: ".name == $var0",
        variables: { $var0: "nginx" },
      } as IPkgRepositoryFilter,
    },
    skipTLS: false,
    passCredentials: true,
    authMethod:
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  } as IPkgRepoFormData;

  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
    jest.restoreAllMocks();
  });

  it("create repository", async () => {
    const createRepoUrl = url.backend.apprepositories.create(cluster, namespace);
    moxios.stubRequest(createRepoUrl, {
      status: 200,
      response: {},
    });

    await AppRepository.create(cluster, namespace, pkgRepoFormData);

    const request = moxios.requests.mostRecent();
    expect(request.config.method).toEqual("post");
    expect(request.url).toBe(createRepoUrl);
    expect(JSON.parse(request.config.data)).toEqual({
      appRepository: {
        authHeader: "repo-authHeader",
        authRegCreds: ["repo-secret1"],
        description: "repo-description",
        filterRule: {
          jq: ".name == $var0",
          variables: {
            $var0: "nginx",
          },
        },
        name: "repo-test",
        ociRepositories: ["oci-repo1"],
        passCredentials: true,
        repoURL: "repo-url",
        syncJobPodTemplate: "",
        tlsInsecureSkipVerify: false,
        type: "repo-type",
        customCA: "repo-customCA",
      },
    });
  });

  it("update repository", async () => {
    const updateRepoUrl = url.backend.apprepositories.update(
      cluster,
      namespace,
      pkgRepoFormData.name,
    );
    moxios.stubRequest(updateRepoUrl, {
      status: 200,
      response: {},
    });

    await AppRepository.update(cluster, namespace, pkgRepoFormData);

    const request = moxios.requests.mostRecent();
    expect(request.config.method).toEqual("put");
    expect(request.url).toBe(updateRepoUrl);
    expect(JSON.parse(request.config.data)).toEqual({
      appRepository: {
        authHeader: "repo-authHeader",
        authRegCreds: ["repo-secret1"],
        description: "repo-description",
        filterRule: {
          jq: ".name == $var0",
          variables: {
            $var0: "nginx",
          },
        },
        name: "repo-test",
        ociRepositories: ["oci-repo1"],
        passCredentials: true,
        repoURL: "repo-url",
        syncJobPodTemplate: "",
        tlsInsecureSkipVerify: false,
        type: "repo-type",
        customCA: "repo-customCA",
      },
    });
  });

  it("delete repository", async () => {
    const deleteRepoUrl = url.backend.apprepositories.delete(
      cluster,
      namespace,
      pkgRepoFormData.name,
    );
    moxios.stubRequest(deleteRepoUrl, {
      status: 200,
      response: {},
    });

    await AppRepository.delete(cluster, namespace, pkgRepoFormData.name);

    const request = moxios.requests.mostRecent();
    expect(request.config.method).toEqual("delete");
    expect(request.url).toBe(deleteRepoUrl);
  });
});
