import context from "jest-plugin-context";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import actions from ".";
import { AppRepository } from "../shared/AppRepository";
import Chart from "../shared/Chart";
import Secret from "../shared/Secret";
import { IAppRepository, NotFoundError } from "../shared/types";

const { repos: repoActions } = actions;
const mockStore = configureMockStore([thunk]);

let store: any;
const appRepo = { spec: { resyncRequests: 10000 } };
const kubeappsNamespace = "kubeapps-namespace";

const safeYAMLTemplate = `
spec:
  containers:
    - env:
      - name: FOO
        value: BAR
`;

beforeEach(() => {
  store = mockStore({
    config: { kubeappsNamespace },
    clusters: {
      currentCluster: "default",
      clusters: {
        default: {
          currentNamespace: kubeappsNamespace,
        },
      },
    },
  });
  AppRepository.list = jest.fn().mockImplementationOnce(() => {
    return { items: { foo: "bar" } };
  });
  AppRepository.delete = jest.fn();
  AppRepository.get = jest.fn().mockImplementationOnce(() => {
    return appRepo;
  });
  AppRepository.update = jest.fn();
  AppRepository.create = jest.fn().mockImplementationOnce(() => {
    return { appRepository: { metadata: { name: "repo-abc" } } };
  });
  Secret.list = jest.fn().mockReturnValue({
    items: [],
  });
});

afterEach(jest.resetAllMocks);

// Regular action creators
interface ITestCase {
  name: string;
  action: (...args: any[]) => any;
  args?: any;
  payload?: any;
}

const repo = { metadata: { name: "my-repo" } } as IAppRepository;

const actionTestCases: ITestCase[] = [
  { name: "addRepo", action: repoActions.addRepo },
  { name: "addedRepo", action: repoActions.addedRepo, args: repo, payload: repo },
  { name: "requestRepos", action: repoActions.requestRepos },
  { name: "receiveRepos", action: repoActions.receiveRepos, args: [[repo]], payload: [repo] },
  { name: "requestRepo", action: repoActions.requestRepo },
  { name: "receiveRepo", action: repoActions.receiveRepo, args: repo, payload: repo },
  { name: "clearRepo", action: repoActions.clearRepo, payload: {} },
  { name: "showForm", action: repoActions.showForm },
  { name: "hideForm", action: repoActions.hideForm },
  { name: "resetForm", action: repoActions.resetForm },
  { name: "submitForm", action: repoActions.submitForm },
  { name: "redirect", action: repoActions.redirect, args: "/foo", payload: "/foo" },
  { name: "redirected", action: repoActions.redirected },
  {
    name: "errorRepos",
    action: repoActions.errorRepos,
    args: [new Error("foo"), "create"],
    payload: { err: new Error("foo"), op: "create" },
  },
];

actionTestCases.forEach(tc => {
  describe(tc.name, () => {
    it("has expected structure", () => {
      const actionResult =
        tc.args && tc.args.length && typeof tc.args === "object"
          ? tc.action.call(null, ...tc.args)
          : tc.action.call(null, tc.args);
      expect(actionResult).toEqual({
        type: getType(tc.action),
        payload: tc.payload,
      });
    });
  });
});

// Async action creators
describe("deleteRepo", () => {
  context("dispatches requestRepos and receivedRepos after deletion if no error", () => {
    const currentNamespace = "current-namespace";
    it("dispatches requestRepos with current namespace", async () => {
      const storeWithFlag: any = mockStore({
        clusters: {
          currentCluster: "defaultCluster",
          clusters: {
            defaultCluster: {
              currentNamespace,
            },
          },
        },
      });
      const expectedActions = [
        {
          type: getType(repoActions.requestRepos),
          payload: currentNamespace,
        },
        {
          type: getType(repoActions.receiveRepos),
          payload: { foo: "bar" },
        },
      ];

      await storeWithFlag.dispatch(repoActions.deleteRepo("foo", "my-namespace"));
      expect(storeWithFlag.getActions()).toEqual(expectedActions);
    });
  });

  it("dispatches errorRepos if error deleting", async () => {
    AppRepository.delete = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "delete" },
      },
    ];

    await store.dispatch(repoActions.deleteRepo("foo", "my-namespace"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("resyncRepo", () => {
  it("dispatches errorRepos if error on #get", async () => {
    AppRepository.get = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "update" },
      },
    ];

    await store.dispatch(repoActions.resyncRepo("foo", "my-namespace"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches errorRepos if error on #update", async () => {
    AppRepository.resync = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "update" },
      },
    ];

    await store.dispatch(repoActions.resyncRepo("foo", "my-namespace"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("resyncAllRepos", () => {
  it("resyncs each repo using its namespace", async () => {
    const appRepoGetMock = jest.fn();
    AppRepository.resync = appRepoGetMock;
    await store.dispatch(
      repoActions.resyncAllRepos([
        {
          name: "foo",
          namespace: "namespace-1",
        },
        {
          name: "bar",
          namespace: "namespace-2",
        },
      ]),
    );

    expect(appRepoGetMock).toHaveBeenCalledTimes(2);
    expect(appRepoGetMock.mock.calls[0]).toEqual(["foo", "namespace-1"]);
    expect(appRepoGetMock.mock.calls[1]).toEqual(["bar", "namespace-2"]);
  });
});

describe("fetchRepos", () => {
  const namespace = "default";
  it("dispatches requestRepos and receivedRepos if no error", async () => {
    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: namespace,
      },
      {
        type: getType(repoActions.receiveRepos),
        payload: { foo: "bar" },
      },
      {
        type: getType(repoActions.receiveReposSecrets),
        payload: [],
      },
    ];

    await store.dispatch(repoActions.fetchRepos(namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("includes secrets that are owned by an apprepo", async () => {
    const appRepoSecret = {
      metadata: {
        name: "foo",
        ownerReferences: [
          {
            kind: "AppRepository",
          },
        ],
      },
    };
    const otherSecret = {
      metadata: {
        name: "bar",
        ownerReferences: [
          {
            kind: "Other",
          },
        ],
      },
    };
    Secret.list = jest.fn().mockReturnValue({
      items: [appRepoSecret, otherSecret],
    });
    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: namespace,
      },
      {
        type: getType(repoActions.receiveRepos),
        payload: { foo: "bar" },
      },
      {
        type: getType(repoActions.receiveReposSecrets),
        payload: [appRepoSecret],
      },
    ];

    await store.dispatch(repoActions.fetchRepos(namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches requestRepos and errorRepos if error fetching", async () => {
    AppRepository.list = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: namespace,
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "fetch" },
      },
    ];

    await store.dispatch(repoActions.fetchRepos(namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("fetches repos from several namespaces and joins them", async () => {
    AppRepository.list = jest
      .fn()
      .mockImplementationOnce(() => {
        return { items: [{ name: "repo1" }] };
      })
      .mockImplementationOnce(() => {
        return { items: [{ name: "repo2" }] };
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: namespace,
      },
      {
        type: getType(repoActions.requestRepos),
        payload: "other-ns",
      },
      {
        type: getType(repoActions.receiveReposSecrets),
        payload: [],
      },
      {
        type: getType(repoActions.receiveRepos),
        payload: [{ name: "repo1" }, { name: "repo2" }],
      },
    ];

    await store.dispatch(repoActions.fetchRepos(namespace, "other-ns"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("installRepo", () => {
  const installRepoCMD = repoActions.installRepo(
    "my-repo",
    "my-namespace",
    "http://foo.bar",
    "",
    "",
    "",
    [],
  );

  context("when authHeader provided", () => {
    const installRepoCMDAuth = repoActions.installRepo(
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "Bearer: abc",
      "",
      "",
      [],
    );

    it("calls AppRepository create including a auth struct", async () => {
      await store.dispatch(installRepoCMDAuth);
      expect(AppRepository.create).toHaveBeenCalledWith(
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "Bearer: abc",
        "",
        {},
        [],
      );
    });

    it("returns true", async () => {
      const res = await store.dispatch(installRepoCMDAuth);
      expect(res).toBe(true);
    });
  });

  context("when a customCA is provided", () => {
    const installRepoCMDAuth = repoActions.installRepo(
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "",
      "This is a cert!",
      "",
      [],
    );

    it("calls AppRepository create including a auth struct", async () => {
      await store.dispatch(installRepoCMDAuth);
      expect(AppRepository.create).toHaveBeenCalledWith(
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "",
        "This is a cert!",
        {},
        [],
      );
    });

    it("returns true", async () => {
      const res = await store.dispatch(installRepoCMDAuth);
      expect(res).toBe(true);
    });

    context("when a pod template is provided", () => {
      it("calls AppRepository create including pod template", async () => {
        await store.dispatch(
          repoActions.installRepo(
            "my-repo",
            "my-namespace",
            "http://foo.bar",
            "",
            "",
            safeYAMLTemplate,
            [],
          ),
        );

        expect(AppRepository.create).toHaveBeenCalledWith(
          "my-repo",
          "my-namespace",
          "http://foo.bar",
          "",
          "",
          {
            spec: { containers: [{ env: [{ name: "FOO", value: "BAR" }] }] },
          },
          [],
        );
      });

      // Example from https://nealpoole.com/blog/2013/06/code-execution-via-yaml-in-js-yaml-nodejs-module/
      const unsafeYAMLTemplate =
        '"toString": !<tag:yaml.org,2002:js/function> "function (){very_evil_thing();}"';

      it("does not call AppRepository create with an unsafe pod template", async () => {
        await store.dispatch(
          repoActions.installRepo(
            "my-repo",
            "my-namespace",
            "http://foo.bar",
            "",
            "",
            unsafeYAMLTemplate,
            [],
          ),
        );
        expect(AppRepository.create).not.toHaveBeenCalled();
      });
    });
  });

  context("when authHeader and customCA are empty", () => {
    it("calls AppRepository create without a auth struct", async () => {
      await store.dispatch(installRepoCMD);
      expect(AppRepository.create).toHaveBeenCalledWith(
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "",
        "",
        {},
        [],
      );
    });

    it("returns true", async () => {
      const res = await store.dispatch(installRepoCMD);
      expect(res).toBe(true);
    });
  });

  it("dispatches addRepo and errorRepos if error fetching", async () => {
    AppRepository.create = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(repoActions.addRepo),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "create" },
      },
    ];

    await store.dispatch(installRepoCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false if error fetching", async () => {
    AppRepository.create = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const res = await store.dispatch(installRepoCMD);
    expect(res).toEqual(false);
  });

  it("dispatches addRepo and addedRepo if no error", async () => {
    const expectedActions = [
      {
        type: getType(repoActions.addRepo),
      },
      {
        type: getType(repoActions.addedRepo),
        payload: { metadata: { name: "repo-abc" } },
      },
    ];

    await store.dispatch(installRepoCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("uses kubeapps own namespace if namespace is _all", async () => {
    await store.dispatch(
      repoActions.installRepo("my-repo", "_all", "http://foo.bar", "", "", "", []),
    );

    expect(AppRepository.create).toHaveBeenCalledWith(
      "my-repo",
      "kubeapps-namespace",
      "http://foo.bar",
      "",
      "",
      {},
      [],
    );
  });

  it("includes registry secrets if given", async () => {
    await store.dispatch(
      repoActions.installRepo("my-repo", "_all", "http://foo.bar", "", "", "", ["repo-1"]),
    );

    expect(AppRepository.create).toHaveBeenCalledWith(
      "my-repo",
      "kubeapps-namespace",
      "http://foo.bar",
      "",
      "",
      {},
      ["repo-1"],
    );
  });
});

describe("updateRepo", () => {
  it("updates a repo with an auth header", async () => {
    const r = {
      metadata: { name: "repo-abc" },
      spec: { auth: { header: { secretKeyRef: { name: "apprepo-repo-abc" } } } },
    };
    const secret = { metadata: { name: "apprepo-repo-abc" } };
    AppRepository.update = jest.fn().mockReturnValue({
      appRepository: r,
    });
    Secret.get = jest.fn().mockReturnValue(secret);
    const expectedActions = [
      {
        type: getType(repoActions.requestRepoUpdate),
      },
      {
        type: getType(repoActions.repoUpdated),
        payload: r,
      },
      {
        type: getType(repoActions.receiveReposSecret),
        payload: secret,
      },
    ];

    await store.dispatch(
      repoActions.updateRepo(
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "foo",
        "bar",
        safeYAMLTemplate,
        ["repo-1"],
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(AppRepository.update).toHaveBeenCalledWith(
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "foo",
      "bar",
      { spec: { containers: [{ env: [{ name: "FOO", value: "BAR" }] }] } },
      ["repo-1"],
    );
  });

  it("updates a repo with an customCA", async () => {
    const r = {
      metadata: { name: "repo-abc" },
      spec: { auth: { customCA: { secretKeyRef: { name: "apprepo-repo-abc" } } } },
    };
    const secret = { metadata: { name: "apprepo-repo-abc" } };
    AppRepository.update = jest.fn().mockReturnValue({
      appRepository: r,
    });
    Secret.get = jest.fn().mockReturnValue(secret);
    const expectedActions = [
      {
        type: getType(repoActions.requestRepoUpdate),
      },
      {
        type: getType(repoActions.repoUpdated),
        payload: r,
      },
      {
        type: getType(repoActions.receiveReposSecret),
        payload: secret,
      },
    ];

    await store.dispatch(
      repoActions.updateRepo(
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "foo",
        "bar",
        safeYAMLTemplate,
        ["repo-1"],
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(AppRepository.update).toHaveBeenCalledWith(
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "foo",
      "bar",
      { spec: { containers: [{ env: [{ name: "FOO", value: "BAR" }] }] } },
      ["repo-1"],
    );
  });

  it("returns an error if failed", async () => {
    AppRepository.update = jest.fn(() => {
      throw new Error("boom");
    });
    const expectedActions = [
      {
        type: getType(repoActions.requestRepoUpdate),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("boom"), op: "update" },
      },
    ];

    await store.dispatch(
      repoActions.updateRepo(
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "foo",
        "bar",
        safeYAMLTemplate,
        [],
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("checkChart", () => {
  it("dispatches requestRepo and receivedRepo if no error", async () => {
    Chart.fetchChartVersions = jest.fn();
    const expectedActions = [
      {
        type: getType(repoActions.requestRepo),
      },
      {
        type: getType(repoActions.receiveRepo),
        payload: appRepo,
      },
    ];

    await store.dispatch(repoActions.checkChart("other-namespace", "my-repo", "my-chart"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(Chart.fetchChartVersions).toBeCalledWith("other-namespace", "my-repo/my-chart");
  });

  it("dispatches requestRepo and errorChart if error fetching", async () => {
    Chart.fetchChartVersions = jest.fn(() => {
      throw new Error();
    });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepo),
      },
      {
        type: getType(actions.charts.errorChart),
        payload: new NotFoundError("Chart my-chart not found in the repository my-repo."),
      },
    ];

    await store.dispatch(repoActions.checkChart("other-namespace", "my-repo", "my-chart"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(Chart.fetchChartVersions).toBeCalledWith("other-namespace", "my-repo/my-chart");
  });
});

describe("validateRepo", () => {
  it("dispatches repoValidating and repoValidated if no error", async () => {
    AppRepository.validate = jest.fn().mockReturnValue({
      code: 200,
      message: "OK",
    });
    const expectedActions = [
      {
        type: getType(repoActions.repoValidating),
      },
      {
        type: getType(repoActions.repoValidated),
        payload: { code: 200, message: "OK" },
      },
    ];

    const res = await store.dispatch(repoActions.validateRepo("url", "auth", "cert"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(res).toBe(true);
  });

  it("dispatches checkRepo and errorRepos when the validation failed", async () => {
    const error = new Error("boom!");
    AppRepository.validate = jest.fn(() => {
      throw error;
    });
    const expectedActions = [
      {
        type: getType(repoActions.repoValidating),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: error, op: "validate" },
      },
    ];
    const res = await store.dispatch(repoActions.validateRepo("url", "auth", "cert"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(res).toBe(false);
  });

  it("dispatches checkRepo and errorRepos when the validation cannot be parsed", async () => {
    AppRepository.validate = jest.fn().mockReturnValue({
      code: 409,
      message: "forbidden",
    });
    const expectedActions = [
      {
        type: getType(repoActions.repoValidating),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: {
          err: new Error('{"code":409,"message":"forbidden"}'),
          op: "validate",
        },
      },
    ];
    const res = await store.dispatch(repoActions.validateRepo("url", "auth", "cert"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(res).toBe(false);
  });
});

describe("fetchImagePullSecrets", () => {
  it("fetches image pull secrets", async () => {
    const secret1 = {
      type: "kubernetes.io/dockerconfigjson",
    };
    const secret2 = {
      type: "Opaque",
    };
    Secret.list = jest.fn().mockReturnValue({
      items: [secret1, secret2],
    });
    const expectedActions = [
      {
        type: getType(repoActions.requestImagePullSecrets),
        payload: "default",
      },
      {
        type: getType(repoActions.receiveImagePullSecrets),
        payload: [secret1],
      },
    ];
    await store.dispatch(repoActions.fetchImagePullSecrets("default"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Secret.list = jest.fn(() => {
      throw new Error("boom");
    });
    const expectedActions = [
      {
        type: getType(repoActions.requestImagePullSecrets),
        payload: "default",
      },
      {
        type: getType(repoActions.errorRepos),
        payload: {
          err: new Error("boom"),
          op: "fetch",
        },
      },
    ];
    await store.dispatch(repoActions.fetchImagePullSecrets("default"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("createDockerRegistrySecret", () => {
  it("creates a docker registry", async () => {
    const secret = {
      type: "kubernetes.io/dockerconfigjson",
    };
    Secret.createPullSecret = jest.fn().mockReturnValue(secret);
    const expectedActions = [
      {
        type: getType(repoActions.createImagePullSecret),
        payload: secret,
      },
    ];
    await store.dispatch(repoActions.createDockerRegistrySecret("", "", "", "", "", ""));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Secret.createPullSecret = jest.fn(() => {
      throw new Error("boom");
    });
    const expectedActions = [
      {
        type: getType(repoActions.errorRepos),
        payload: {
          err: new Error("boom"),
          op: "fetch",
        },
      },
    ];
    await store.dispatch(repoActions.createDockerRegistrySecret("", "", "", "", "", ""));
    expect(store.getActions()).toEqual(expectedActions);
  });
});
