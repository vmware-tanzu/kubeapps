// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { PackageRepositorySummary } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as ReactRedux from "react-redux";
import { IPackageRepositoryState } from "reducers/repos";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import SelectRepoForm from "./SelectRepoForm";

const defaultContext = {
  cluster: "default-cluster",
  namespace: "default",
};

const installedPackageDetail = {
  availablePackageRef: {
    context: defaultContext,
    identifier: "bitnami/my-package",
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
} as InstalledPackageDetail;

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.operators };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    fetchRepoSummaries: jest.fn(),
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
  actions.repos = { ...actions.repos, fetchRepoSummaries: fetch };
  const props = {
    cluster: defaultContext.cluster,
    namespace: initialState.config.kubeappsNamespace, // global
    app: installedPackageDetail,
  };
  mountWrapper(defaultStore, <SelectRepoForm {...props} />);
  expect(fetch).toHaveBeenCalledWith(initialState.config.kubeappsNamespace, true);
});

it("should fetch repositories", () => {
  const fetch = jest.fn();
  actions.repos = { ...actions.repos, fetchRepoSummaries: fetch };
  mountWrapper(defaultStore, <SelectRepoForm {...defaultContext} />);
  expect(fetch).toHaveBeenCalledWith(defaultContext.namespace, true);
});

it("should render a loading page if fetching", () => {
  expect(
    mountWrapper(
      getStore({ repos: { isFetching: true } } as Partial<IStoreState>),
      <SelectRepoForm {...defaultContext} />,
    ).find("LoadingWrapper"),
  ).toExist();
});

it("render an error if failed to request repos", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { fetch: new Error("boom") } } } as Partial<IStoreState>),
    <SelectRepoForm {...defaultContext} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom");
});

it("render a warning if there are no repos", () => {
  const wrapper = mountWrapper(defaultStore, <SelectRepoForm {...defaultContext} />);
  expect(wrapper.find(Alert)).toIncludeText("Repositories not found");
});

it("should select a repo", () => {
  const findPackageInRepo = jest.fn();
  actions.repos = { ...actions.repos, findPackageInRepo };
  const repo = {
    name: "bitnami",
    url: "http://repo",
    packageRepoRef: {
      context: { namespace: "default", cluster: "default" },
      identifier: "bitnami",
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    },
  } as PackageRepositorySummary;

  const props = { ...defaultContext, app: installedPackageDetail };
  const wrapper = mountWrapper(
    getStore({
      repos: { reposSummaries: [repo] } as IPackageRepositoryState,
    } as Partial<IStoreState>),
    <SelectRepoForm {...props} />,
  );
  wrapper.find("select").simulate("change", { target: { value: "default/bitnami" } });
  (wrapper.find(CdsButton).prop("onClick") as any)();
  expect(findPackageInRepo).toHaveBeenCalledWith(
    initialState.config.kubeappsCluster,
    repo.packageRepoRef?.context?.namespace,
    repo.name,
    installedPackageDetail,
  );
});
