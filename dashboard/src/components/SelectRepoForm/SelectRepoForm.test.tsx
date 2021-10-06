import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IAppRepository } from "shared/types";
import SelectRepoForm from "./SelectRepoForm";

const defaultProps = {
  cluster: "default",
  namespace: "default",
};

const installedPackageDetail = {
  availablePackageRef: {
    context: { cluster: "default", namespace: "default" },
    identifier: "bitnami/my-package",
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
} as InstalledPackageDetail;

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.operators };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    fetchRepos: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("should fetch only the global repository", () => {
  const fetch = jest.fn();
  actions.repos = { ...actions.repos, fetchRepos: fetch };
  const props = {
    cluster: defaultProps.cluster,
    namespace: initialState.config.kubeappsNamespace, // global
    app: installedPackageDetail,
  };
  mountWrapper(defaultStore, <SelectRepoForm {...props} />);
  expect(fetch).toHaveBeenCalledWith(initialState.config.kubeappsNamespace);
});

it("should fetch repositories", () => {
  const fetch = jest.fn();
  actions.repos = { ...actions.repos, fetchRepos: fetch };
  mountWrapper(defaultStore, <SelectRepoForm {...defaultProps} />);
  expect(fetch).toHaveBeenCalledWith(defaultProps.namespace, true);
});

it("should render a loading page if fetching", () => {
  expect(
    mountWrapper(
      getStore({ repos: { isFetching: true } }),
      <SelectRepoForm {...defaultProps} />,
    ).find("LoadingWrapper"),
  ).toExist();
});

it("render an error if failed to request repos", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { fetch: new Error("boom") } } }),
    <SelectRepoForm {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom");
});

it("render a warning if there are no repos", () => {
  const wrapper = mountWrapper(defaultStore, <SelectRepoForm {...defaultProps} />);
  expect(wrapper.find(Alert)).toIncludeText("Repositories not found");
});

it("should select a repo", () => {
  const findPackageInRepo = jest.fn();
  actions.repos = { ...actions.repos, findPackageInRepo };
  const repo = {
    metadata: {
      name: "bitnami",
      namespace: "default",
    },
    spec: {
      url: "http://repo",
    },
  } as IAppRepository;

  const props = { ...defaultProps, app: installedPackageDetail };
  const wrapper = mountWrapper(
    getStore({ repos: { repos: [repo] } }),
    <SelectRepoForm {...props} />,
  );
  wrapper.find("select").simulate("change", { target: { value: "default/bitnami" } });
  (wrapper.find(CdsButton).prop("onClick") as any)();
  expect(findPackageInRepo).toHaveBeenCalledWith(
    initialState.config.kubeappsCluster,
    repo.metadata.namespace,
    repo.metadata.name,
    installedPackageDetail,
  );
});
