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

beforeEach(() => {
  store = mockStore({
    config: { namespace: kubeappsNamespace },
    namespace: { current: kubeappsNamespace },
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
  Secret.create = jest.fn();
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
  context("dispatches requestRepos and receivedRepos after deletion if no error", async () => {
    const currentNamespace = "current-namespace";
    it("dispatches requestRepos with kubeapps namespace when reposPerNamespace is not set", async () => {
      const storeWithFlag: any = mockStore({
        config: {
          namespace: kubeappsNamespace,
          featureFlags: { reposPerNamespace: false },
        },
        namespace: { current: currentNamespace },
      });
      const expectedActions = [
        {
          type: getType(repoActions.requestRepos),
          payload: kubeappsNamespace,
        },
        {
          type: getType(repoActions.receiveRepos),
          payload: { foo: "bar" },
        },
      ];

      await storeWithFlag.dispatch(repoActions.deleteRepo("foo", "my-namespace"));
      expect(storeWithFlag.getActions()).toEqual(expectedActions);
    });

    it("dispatches requestRepos with current namespace when reposPerNamespace is set", async () => {
      const storeWithFlag: any = mockStore({
        config: {
          namespace: kubeappsNamespace,
          featureFlags: { reposPerNamespace: true },
        },
        namespace: { current: currentNamespace },
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
    AppRepository.update = jest.fn().mockImplementationOnce(() => {
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
    AppRepository.get = appRepoGetMock;
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
});

describe("installRepo", () => {
  const installRepoCMD = repoActions.installRepo(
    "my-repo",
    "my-namespace",
    "http://foo.bar",
    "",
    "",
    "",
  );

  context("when authHeader provided", () => {
    const installRepoCMDAuth = repoActions.installRepo(
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "Bearer: abc",
      "",
      "",
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
      );
    });

    it("does not create the K8s secret as API includes this", async () => {
      await store.dispatch(installRepoCMDAuth);
      expect(Secret.create).not.toHaveBeenCalled();
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
      );
    });

    it("does not create the K8s secret as API includes this", async () => {
      await store.dispatch(installRepoCMDAuth);
      expect(Secret.create).not.toHaveBeenCalled();
    });

    it("returns true", async () => {
      const res = await store.dispatch(installRepoCMDAuth);
      expect(res).toBe(true);
    });

    context("when a pod template is provided", () => {
      const safeYAMLTemplate = `
spec:
  containers:
    - env:
      - name: FOO
        value: BAR
`;

      it("calls AppRepository create including pod template", async () => {
        await store.dispatch(
          repoActions.installRepo(
            "my-repo",
            "my-namespace",
            "http://foo.bar",
            "",
            "",
            safeYAMLTemplate,
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
    await store.dispatch(repoActions.installRepo("my-repo", "_all", "http://foo.bar", "", "", ""));

    expect(AppRepository.create).toHaveBeenCalledWith(
      "my-repo",
      "kubeapps-namespace",
      "http://foo.bar",
      "",
      "",
      {},
    );
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

    await store.dispatch(repoActions.checkChart(kubeappsNamespace, "my-repo", "my-chart"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(Chart.fetchChartVersions).toBeCalledWith("kubeapps-namespace", "my-repo/my-chart");
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

    await store.dispatch(repoActions.checkChart(kubeappsNamespace, "my-repo", "my-chart"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(Chart.fetchChartVersions).toBeCalledWith("kubeapps-namespace", "my-repo/my-chart");
  });
});

describe("validateRepo", () => {
  it("dispatches repoValidating and repoValidated if no error", async () => {
    AppRepository.validate = jest.fn(() => "OK");
    const expectedActions = [
      {
        type: getType(repoActions.repoValidating),
      },
      {
        type: getType(repoActions.repoValidated),
        payload: "OK",
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
    AppRepository.validate = jest.fn(() => {
      return { statusCode: 409 };
    });
    const expectedActions = [
      {
        type: getType(repoActions.repoValidating),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: {
          err: new Error('Unable to parse validation response, got: {"statusCode":409}'),
          op: "validate",
        },
      },
    ];
    const res = await store.dispatch(repoActions.validateRepo("url", "auth", "cert"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(res).toBe(false);
  });
});
