// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import * as moxios from "moxios";
import { axiosWithAuth } from "./AxiosInstance";

describe("RepositoriesService", () => {
  // const cluster = "cluster";
  // const namespace = "namespace";
  // const plugin: Plugin = { name: "my.plugin", version: "0.0.1" };

  // const repo = {
  //   name: "repo-test",
  //   type: "repo-type",
  //   description: "repo-description",
  //   authHeader: "repo-authHeader",
  //   customCA: "repo-customCA",
  //   passCredentials: true,
  //   authMethod:
  //     PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  //   basicAuth: {
  //     username: "user",
  //     password: "password",
  //   },
  //   skipTLS: false,
  //   tlsCertKey: {
  //     cert: "",
  //     key: "",
  //   },
  //   sshCreds: {
  //     knownHosts: "",
  //     privateKey: "",
  //   },
  //   secretTLSName: "",

  //   opaqueCreds: {
  //     data: {},
  //   },
  //   url: "repo-url",
  //   secretAuthName: "repo-secret1",
  //   dockerRegCreds: {
  //     password: "",
  //     username: "",
  //     email: "",
  //     server: "",
  //   },
  //   interval: 3600,
  //   plugin,
  //   customDetails: {
  //     ociRepositories: ["oci-repo1"],
  //     dockerRegistrySecrets: ["repo-authRegCreds"],
  //     filterRule: {
  //       jq: ".name == $var0",
  //       variables: { $var0: "nginx" },
  //     },
  //   },
  // } as IPkgRepoFormData;

  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
    jest.restoreAllMocks();
  });

  // TODO(agamez): add tests back
  // it("create repository", async () => {
  //   const createRepoUrl = url.backend.apprepositories.create(cluster, namespace);
  //   moxios.stubRequest(createRepoUrl, {
  //     status: 200,
  //     response: {},
  //   });

  //   await PackageRepositoriesService.addPackageRepository(
  //     cluster,
  //     repo.name,
  //     plugin,
  //     namespace,
  //     repo.repoURL,
  //     repo.type,
  //     repo.description,
  //     repo.authHeader,
  //     repo.authRegCreds,
  //     repo.customCA,
  //     repo.registrySecrets,
  //     repo.ociRepositories,
  //     repo.tlsInsecureSkipVerify,
  //     repo.passCredentials,
  //     true,
  //     repo.authMethod,
  //     repo.interval,
  //     repo.username,
  //     repo.password,
  //     false,
  //     repo.filterRule,
  //   );

  //   const request = moxios.requests.mostRecent();
  //   expect(request.config.method).toEqual("post");
  //   expect(request.url).toBe(createRepoUrl);
  //   expect(JSON.parse(request.config.data)).toEqual({ appRepository: repo });
  // });

  // it("update repository", async () => {
  //   const updateRepoUrl = url.backend.apprepositories.update(cluster, namespace, repo.name);
  //   moxios.stubRequest(updateRepoUrl, {
  //     status: 200,
  //     response: {},
  //   });

  //   await PackageRepositoriesService.updatePackageRepository(
  //     cluster,
  //     repo.name,
  //     plugin,
  //     namespace,
  //     repo.repoURL,
  //     repo.type,
  //     repo.description,
  //     repo.authHeader,
  //     repo.authRegCreds,
  //     repo.customCA,
  //     repo.registrySecrets,
  //     repo.ociRepositories,
  //     repo.tlsInsecureSkipVerify,
  //     repo.passCredentials,
  //     repo.authMethod,
  //     repo.interval,
  //     repo.username,
  //     repo.password,
  //     false,
  //     repo.filterRule,
  //   );

  //   const request = moxios.requests.mostRecent();
  //   expect(request.config.method).toEqual("put");
  //   expect(request.url).toBe(updateRepoUrl);
  //   expect(JSON.parse(request.config.data)).toEqual({ appRepository: repo });
  // });

  // it("delete repository", async () => {
  //   const deleteRepoUrl = url.backend.apprepositories.delete(cluster, namespace, repo.name);
  //   moxios.stubRequest(deleteRepoUrl, {
  //     status: 200,
  //     response: {},
  //   });

  //   await PackageRepositoriesService.deletePackageRepository({
  //     identifier: repo.name,
  //     context: { cluster, namespace },
  //     plugin,
  //   } as PackageRepositoryReference);

  //   const request = moxios.requests.mostRecent();
  //   expect(request.config.method).toEqual("delete");
  //   expect(request.url).toBe(deleteRepoUrl);
  // });
});
