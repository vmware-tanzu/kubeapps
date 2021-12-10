import * as moxios from "moxios";
import { AppRepository } from "./AppRepository";
import { axiosWithAuth } from "./AxiosInstance";
import * as url from "./url";
import { IAppRepositoryFilter } from "shared/types";

describe("AppRepository", () => {
  const cluster = "cluster";
  const namespace = "namespace";
  const repo = {
    name: "repo-test",
    repoURL: "repo-url",
    type: "repo-type",
    description: "repo-description",
    authHeader: "repo-authHeader",
    authRegCreds: "repo-authRegCreds",
    customCA: "repo-customCA",
    syncJobPodTemplate: { type: "helm" },
    registrySecrets: ["repo-secret1"],
    ociRepositories: ["oci-repo1"],
    tlsInsecureSkipVerify: false,
    passCredentials: true,
    filterRule: {
      jq: ".name == $var0",
      variables: { $var0: "nginx" },
    } as IAppRepositoryFilter,
  };

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

    await AppRepository.create(
      cluster,
      repo.name,
      namespace,
      repo.repoURL,
      repo.type,
      repo.description,
      repo.authHeader,
      repo.authRegCreds,
      repo.customCA,
      repo.syncJobPodTemplate,
      repo.registrySecrets,
      repo.ociRepositories,
      repo.tlsInsecureSkipVerify,
      repo.passCredentials,
      repo.filterRule,
    );

    const request = moxios.requests.mostRecent();
    expect(request.config.method).toEqual("post");
    expect(request.url).toBe(createRepoUrl);
    expect(JSON.parse(request.config.data)).toEqual({ appRepository: repo });
  });

  it("update repository", async () => {
    const updateRepoUrl = url.backend.apprepositories.update(cluster, namespace, repo.name);
    moxios.stubRequest(updateRepoUrl, {
      status: 200,
      response: {},
    });

    await AppRepository.update(
      cluster,
      repo.name,
      namespace,
      repo.repoURL,
      repo.type,
      repo.description,
      repo.authHeader,
      repo.authRegCreds,
      repo.customCA,
      repo.syncJobPodTemplate,
      repo.registrySecrets,
      repo.ociRepositories,
      repo.tlsInsecureSkipVerify,
      repo.passCredentials,
      repo.filterRule,
    );

    const request = moxios.requests.mostRecent();
    expect(request.config.method).toEqual("put");
    expect(request.url).toBe(updateRepoUrl);
    expect(JSON.parse(request.config.data)).toEqual({ appRepository: repo });
  });

  it("delete repository", async () => {
    const deleteRepoUrl = url.backend.apprepositories.delete(cluster, namespace, repo.name);
    moxios.stubRequest(deleteRepoUrl, {
      status: 200,
      response: {},
    });

    await AppRepository.delete(cluster, namespace, repo.name);

    const request = moxios.requests.mostRecent();
    expect(request.config.method).toEqual("delete");
    expect(request.url).toBe(deleteRepoUrl);
  });
});
