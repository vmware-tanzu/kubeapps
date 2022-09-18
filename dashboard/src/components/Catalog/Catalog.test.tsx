// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { deepClone } from "@cds/core/internal/utils/identity";
import actions from "actions";
import FilterGroup from "components/FilterGroup/FilterGroup";
import InfoCard from "components/InfoCard/InfoCard";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper";
import { AvailablePackageSummary, Context } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  PackageRepositoryDetail,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { createMemoryHistory } from "history";
import React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import * as ReactRouter from "react-router";
import { MemoryRouter, Route, Router } from "react-router-dom";
import { IConfigState } from "reducers/config";
import { IOperatorsState } from "reducers/operators";
import { IPackageRepositoryState } from "reducers/repos";
import { getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import {
  IClusterServiceVersion,
  IPackageState,
  IStoreState,
  PluginNames,
} from "../../shared/types";
import SearchFilter from "../SearchFilter/SearchFilter";
import Catalog, { filterNames } from "./Catalog";
import CatalogItems from "./CatalogItems";
import PackageCatalogItem from "./PackageCatalogItem";

const defaultPackageState = {
  isFetching: false,
  hasFinishedFetching: false,
  selected: {} as IPackageState["selected"],
  items: [],
  categories: [],
  nextPageToken: "",
  size: 20,
} as IPackageState;
const defaultProps = {
  cluster: initialState.config.kubeappsCluster,
  namespace: "kubeapps",
  kubeappsNamespace: "kubeapps",
};
const availablePkgSummary1: AvailablePackageSummary = {
  name: "foo",
  categories: [""],
  displayName: "foo",
  iconUrl: "",
  latestVersion: { appVersion: "v1.0.0", pkgVersion: "" },
  shortDescription: "",
  availablePackageRef: {
    identifier: "foo/foo",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
};
const availablePkgSummary2: AvailablePackageSummary = {
  name: "bar",
  categories: ["Database"],
  displayName: "bar",
  iconUrl: "",
  latestVersion: { appVersion: "v2.0.0", pkgVersion: "" },
  shortDescription: "",
  availablePackageRef: {
    identifier: "bar/bar",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
};

const availablePkgSummary3: AvailablePackageSummary = {
  ...availablePkgSummary2,
  availablePackageRef: {
    identifier: "bar/bar2",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: { name: PluginNames.PACKAGES_KAPP, version: "0.0.1" } as Plugin,
  },
};

const csv = {
  metadata: {
    name: "test-csv",
  },
  spec: {
    provider: {
      name: "me",
    },
    icon: [{ base64data: "data", mediatype: "img/png" }],
    customresourcedefinitions: {
      owned: [
        {
          name: "foo-cluster",
          displayName: "foo-cluster",
          version: "v1.0.0",
          description: "a meaningful description",
        },
      ],
    },
  },
} as IClusterServiceVersion;

const defaultState = {
  packages: defaultPackageState,
  operators: { csvs: [] } as Partial<IOperatorsState>,
  repos: {
    reposSummaries: [],
    isFetching: false,
    repoDetail: {} as PackageRepositoryDetail,
    errors: [],
  } as IPackageRepositoryState,
  config: {
    kubeappsCluster: defaultProps.cluster,
    kubeappsNamespace: defaultProps.kubeappsNamespace,
  } as IConfigState,
} as IStoreState;

const populatedPackageState = {
  ...defaultPackageState,
  items: [availablePkgSummary1, availablePkgSummary2],
} as IStoreState["packages"];

const populatedState = {
  ...defaultState,
  packages: populatedPackageState,
  operators: {
    ...defaultState.operators,
    csvs: [csv],
  },
  config: {
    ...defaultState.config,
    configuredPlugins: [
      { name: PluginNames.PACKAGES_KAPP, version: "0.0.1" },
      { name: "my.plugin", version: "0.0.1" },
    ],
  },
} as IStoreState;

let spyOnUseDispatch: jest.SpyInstance;
let spyOnUseHistory: jest.SpyInstance;

beforeEach(() => {
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  spyOnUseHistory = jest
    .spyOn(ReactRouter, "useHistory")
    .mockReturnValue({ push: jest.fn() } as any);
});

afterEach(() => {
  jest.restoreAllMocks();
  spyOnUseDispatch.mockRestore();
  spyOnUseHistory.mockRestore();
});

const routePathParam = `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/catalog`;
const routePath = "/c/:cluster/ns/:namespace/catalog";
const history = createMemoryHistory({ initialEntries: [routePathParam] });

it("retrieves csvs in the namespace if operators enabled", () => {
  const getCSVs = jest.fn();
  actions.operators.getCSVs = getCSVs;
  const state = deepClone(populatedState) as IStoreState;
  state.config.featureFlags = { operators: true };

  mountWrapper(
    getStore(state),
    <Router history={history}>
      <Route path={routePath}>
        <Catalog />
      </Route>
    </Router>,
  );

  expect(getCSVs).toHaveBeenCalledWith(defaultProps.cluster, defaultProps.namespace);
});

it("not retrieveing csvs in the namespace if operators deactivated", () => {
  const getCSVs = jest.fn();
  actions.operators.getCSVs = getCSVs;
  const state = deepClone(populatedState) as IStoreState;
  state.config.featureFlags = { operators: false };

  mountWrapper(
    getStore(state),
    <Router history={history}>
      <Route path={routePath}>
        <Catalog />
      </Route>
    </Router>,
  );

  expect(getCSVs).not.toHaveBeenCalled();
});

it("shows all the elements", () => {
  const wrapper = mountWrapper(getStore(populatedState), <Catalog />);
  expect(wrapper.find(InfoCard)).toHaveLength(3);
});

it("should not render a message if there are no elements in the catalog but the fetching hasn't ended", () => {
  const wrapper = mountWrapper(getStore(defaultState), <Catalog />);
  const message = wrapper.find(".empty-catalog");
  expect(message).not.toExist();
  expect(message).not.toIncludeText("The current catalog is empty");
});

it("should render a message if there are no elements in the catalog and the fetching has ended", () => {
  const wrapper = mountWrapper(
    getStore({
      ...defaultState,
      packages: { hasFinishedFetching: true },
    } as Partial<IStoreState>),
    <Catalog />,
  );
  wrapper.setProps({ searchFilter: "" });
  const message = wrapper.find(".empty-catalog");
  expect(message).toExist();
  expect(message).toIncludeText("The current catalog is empty");
});

it("should render a spinner if there are no elements but it's still fetching", () => {
  const wrapper = mountWrapper(
    getStore({ ...defaultState, packages: { hasFinishedFetching: false } } as Partial<IStoreState>),
    <Catalog />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("should not render a spinner if there are no elements and it finished fetching", () => {
  const wrapper = mountWrapper(
    getStore({ ...defaultState, packages: { hasFinishedFetching: true } } as Partial<IStoreState>),
    <Catalog />,
  );
  expect(wrapper.find(LoadingWrapper)).not.toExist();
});

it("should render a spinner if there already pending elements", () => {
  const wrapper = mountWrapper(
    getStore({
      ...populatedState,
      packages: { hasFinishedFetching: false },
    } as Partial<IStoreState>),
    <Catalog />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("should not render a message if only operators are selected", () => {
  const wrapper = mountWrapper(
    getStore({
      ...populatedState,
      packages: { hasFinishedFetching: true },
    } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam + "?Operators=bar"]}>
      <Route path={routePath}>
        <Catalog />
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(LoadingWrapper)).not.toExist();
});

it("should not render a message if there are no more elements", () => {
  const wrapper = mountWrapper(
    getStore({
      ...populatedState,
      packages: { hasFinishedFetching: true },
    } as Partial<IStoreState>),
    <Catalog />,
  );
  const message = wrapper.find(".end-page-message");
  expect(message).not.toExist();
});

it("should not render a message if there are no more elements but it's searching", () => {
  const wrapper = mountWrapper(
    getStore({
      ...populatedState,
      packages: { hasFinishedFetching: true },
    } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam + "?Search=bar"]}>
      <Route path={routePath}>
        <Catalog />
      </Route>
    </MemoryRouter>,
  );
  const message = wrapper.find(".end-page-message");
  expect(message).not.toExist();
});

it("should render the scroll handler if not finished", () => {
  const wrapper = mountWrapper(
    getStore({
      ...populatedState,
      packages: { hasFinishedFetching: false },
    } as Partial<IStoreState>),
    <Catalog />,
  );
  const scroll = wrapper.find(".scroll-handler");
  expect(scroll).toExist();
  expect(scroll).toHaveProperty("ref");
});

it("should not render the scroll handler if finished", () => {
  const wrapper = mountWrapper(
    getStore({
      ...populatedState,
      packages: { hasFinishedFetching: true },
    } as Partial<IStoreState>),
    <Catalog />,
  );
  const scroll = wrapper.find(".scroll-handler");
  expect(scroll).not.toExist();
});

it("should render an error if it exists", () => {
  const packages = {
    ...defaultPackageState,
    selected: {
      error: new Error("Boom!"),
    },
  } as any;
  const wrapper = mountWrapper(
    getStore({ ...populatedState, packages: packages } as IStoreState),
    <Catalog />,
  );
  const error = wrapper.find(Alert);
  expect(error.prop("theme")).toBe("danger");
  expect(error).toIncludeText("Boom!");
});

it("behaves like a loading wrapper", () => {
  const packages = { isFetching: true, items: [], categories: [], selected: {} } as any;
  const wrapper = mountWrapper(
    getStore({ ...populatedState, packages: packages } as IStoreState),
    <Catalog />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("transforms the received '__' in query params into a ','", () => {
  const wrapper = mountWrapper(
    getStore(populatedState),
    <MemoryRouter initialEntries={[routePathParam + "?Provider=Lightbend__%20Inc."]}>
      <Route path={routePath}>
        <Catalog />
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(".label-info").text()).toBe("Provider: Lightbend, Inc. ");
});

describe("filters by the searched item", () => {
  let spyOnUseDispatch: jest.SpyInstance;
  let spyOnUseEffect: jest.SpyInstance;

  afterEach(() => {
    spyOnUseDispatch.mockRestore();
    spyOnUseEffect.mockRestore();
  });

  it("filters modifying the search box", () => {
    const fetchAvailablePackageSummaries = jest.fn();
    actions.availablepackages.fetchAvailablePackageSummaries = fetchAvailablePackageSummaries;
    const mockDispatch = jest.fn();
    const mockUseEffect = jest.fn();

    spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
    spyOnUseEffect = jest.spyOn(React, "useEffect").mockReturnValue(mockUseEffect as any);

    const wrapper = mountWrapper(
      getStore(populatedState),
      <MemoryRouter initialEntries={[routePathParam + "?Search=bar"]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    act(() => {
      (wrapper.find(SearchFilter).prop("onChange") as any)("bar");
    });
    wrapper.update();
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/kubeapps/catalog?Search=bar"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
  });
});

describe("filters by application type", () => {
  let spyOnUseDispatch: jest.SpyInstance;
  const mockDispatch = jest.fn();

  beforeEach(() => {
    spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  });

  afterEach(() => {
    spyOnUseDispatch.mockRestore();
    mockDispatch.mockRestore();
  });

  it("doesn't show the filter if there are no csvs", () => {
    const wrapper = mountWrapper(getStore(defaultState), <Catalog />);
    expect(
      wrapper.find(FilterGroup).findWhere(g => g.prop("name") === filterNames.TYPE),
    ).not.toExist();
  });

  it("filters only packages", () => {
    const wrapper = mountWrapper(
      getStore(populatedState),
      <MemoryRouter initialEntries={[routePathParam + "?Type=Packages"]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(2);
  });

  it("push filter for only packages", () => {
    const wrapper = mountWrapper(
      getStore(populatedState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Packages");
    expect(input).toHaveLength(1);
    input.simulate("change", { target: { value: "Packages", checked: true } });

    // It should have pushed with the filter
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/kubeapps/catalog?Type=Packages"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
  });

  it("filters only operators", () => {
    const wrapper = mountWrapper(
      getStore(populatedState),
      <MemoryRouter initialEntries={[routePathParam + "?Type=Operators"]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("filters a package type", () => {
    const packages = {
      ...defaultPackageState,
      items: [availablePkgSummary1, availablePkgSummary2, availablePkgSummary3],
    };
    const wrapper = mountWrapper(
      getStore({ ...populatedState, packages: packages } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam + "?Type=Packages&Plugin=Carvel%20Packages"]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("push filter for only operators", () => {
    const wrapper = mountWrapper(
      getStore(populatedState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Operators");
    expect(input).toHaveLength(1);
    input.simulate("change", { target: { value: "Operators", checked: true } });

    // It should have pushed with the filter
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/kubeapps/catalog?Type=Operators"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
  });
});

describe("pagination and package fetching", () => {
  let spyOnUseRef: jest.SpyInstance;
  const refFalse = { current: {} };
  const refTrue = { current: {} };
  Object.defineProperty(refFalse, "current", {
    set(_current) {
      // do nothing
    },
    get() {
      return false;
    },
  });
  Object.defineProperty(refTrue, "current", {
    set(_current) {
      // do nothing
    },
    get() {
      return true;
    },
  });

  beforeEach(() => {
    spyOnUseRef = jest.spyOn(React, "useRef").mockReturnValue(refFalse);
  });
  afterEach(() => {
    spyOnUseRef.mockRestore();
  });

  it("sets the initial state page to 0 before fetching packages", () => {
    const fetchAvailablePackageSummaries = jest.fn();
    actions.availablepackages.fetchAvailablePackageSummaries = fetchAvailablePackageSummaries;

    const packages = {
      ...defaultPackageState,
      hasFinishedFetching: false,
      isFetching: false,
      items: [],
    } as any;
    const wrapper = mountWrapper(
      getStore({ ...populatedState, packages: packages } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );

    expect(wrapper.find(CatalogItems).prop("isFirstPage")).toBe(false);
    expect(wrapper.find(PackageCatalogItem).length).toBe(0);
    expect(fetchAvailablePackageSummaries).toHaveBeenNthCalledWith(
      1,
      "default-cluster",
      "kubeapps",
      "",
      "",
      20,
      "",
    );
  });

  it("avoids re-fetching if isFetching=true", () => {
    jest.useFakeTimers();
    const fetchAvailablePackageSummaries = jest.fn();
    actions.availablepackages.fetchAvailablePackageSummaries = fetchAvailablePackageSummaries;
    spyOnUseRef = jest.spyOn(React, "useRef").mockReturnValue(refTrue);

    const packages = {
      ...defaultPackageState,
      hasFinishedFetching: false,
      isFetching: true,
      items: [availablePkgSummary1],
    } as any;
    const wrapper = mountWrapper(
      getStore({ ...populatedState, packages: packages } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    jest.advanceTimersByTime(2000);

    expect(wrapper.find(CatalogItems).prop("isFirstPage")).toBe(true);
    expect(fetchAvailablePackageSummaries).not.toBeCalled();
  });

  it("disables the filtergroups when isFetching", () => {
    const packages = {
      ...defaultPackageState,
      hasFinishedFetching: true,
      isFetching: true,
      items: [availablePkgSummary1, availablePkgSummary2],
    } as any;
    const wrapper = mountWrapper(
      getStore({ ...populatedState, packages: packages } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );

    wrapper
      .find(FilterGroup)
      .find("input")
      .forEach(i => expect(i.prop("disabled")).toBe(true));
  });

  it("items are translated to CatalogItems after fetching packages", () => {
    const fetchAvailablePackageSummaries = jest.fn();
    actions.availablepackages.fetchAvailablePackageSummaries = fetchAvailablePackageSummaries;

    const packages = {
      ...defaultPackageState,
      hasFinishedFetching: true,
      isFetching: false,
      items: [availablePkgSummary1, availablePkgSummary2],
      nextPageToken: "nextPageToken",
    } as IPackageState;
    const wrapper = mountWrapper(
      getStore({ ...populatedState, packages: packages } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );

    expect(wrapper.find(PackageCatalogItem).length).toBe(2);
  });

  it("does not fetch again after finishing pagination", () => {
    const fetchAvailablePackageSummaries = jest.fn();
    actions.availablepackages.fetchAvailablePackageSummaries = fetchAvailablePackageSummaries;

    const packages = {
      ...defaultPackageState,
      hasFinishedFetching: true,
      isFetching: false,
      items: [availablePkgSummary1],
    } as any;
    const wrapper = mountWrapper(
      getStore({ ...populatedState, packages: packages } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );

    expect(wrapper.find(CatalogItems).prop("isFirstPage")).toBe(false);
    expect(wrapper.find(PackageCatalogItem).length).toBe(1);
    expect(fetchAvailablePackageSummaries).not.toHaveBeenCalled();
  });

  describe("reset", () => {
    const mockDispatch = jest.fn();
    let spyOnUseDispatch: jest.SpyInstance;
    let resetAvailablePackageSummaries: jest.SpyInstance;
    beforeEach(() => {
      spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
      resetAvailablePackageSummaries = jest
        .spyOn(actions.availablepackages, "resetAvailablePackageSummaries")
        .mockImplementation();
    });
    afterEach(() => {
      spyOnUseDispatch.mockRestore();
    });

    it("does not reset during the initial page render", () => {
      const packages = {
        ...defaultPackageState,
        hasFinishedFetching: false,
        isFetching: false,
        items: [],
      } as any;

      mountWrapper(
        getStore({ ...populatedState, packages: packages } as IStoreState),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <Catalog />
          </Route>
        </MemoryRouter>,
      );

      expect(resetAvailablePackageSummaries).not.toHaveBeenCalledWith();
    });

    it("resets the package state when unmounted", () => {
      const packages = {
        ...defaultPackageState,
        hasFinishedFetching: false,
        isFetching: false,
        items: [],
      } as any;

      const wrapper = mountWrapper(
        getStore({ ...populatedState, packages: packages } as IStoreState),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <Catalog />
          </Route>
        </MemoryRouter>,
      );
      wrapper.unmount();

      expect(resetAvailablePackageSummaries).toHaveBeenCalledWith();
    });
    // TODO(agamez): add a test case covering it "resets page when one of the filters changes"
    // https://github.com/vmware-tanzu/kubeapps/pull/2264/files/0d3c77448543668255809bf05039aca704cf729f..22343137efb1c2292b0aa4795f02124306cb055e#r565486271
  });
});

describe("filters by package repository", () => {
  const mockDispatch = jest.fn();
  let spyOnUseDispatch: jest.SpyInstance;
  let fetchRepos: jest.SpyInstance;

  beforeEach(() => {
    spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
    // Can't just assign a mock fn to actions.repos.fetchRepos because it is (correctly) exported
    // as a const fn.
    fetchRepos = jest.spyOn(actions.repos, "fetchRepoSummaries").mockImplementation(() => {
      return jest.fn();
    });
  });

  afterEach(() => {
    mockDispatch.mockRestore();
    spyOnUseDispatch.mockRestore();
    fetchRepos.mockRestore();
  });

  it("doesn't show the filter if there are no apps", () => {
    const wrapper = mountWrapper(
      getStore(defaultState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(
      wrapper.find(FilterGroup).findWhere(g => g.prop("name") === filterNames.REPO),
    ).not.toExist();
  });

  it("filters by repo", () => {
    const wrapper = mountWrapper(
      getStore(populatedState),
      <MemoryRouter initialEntries={[routePathParam + "?Repository=foo"]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("push filter for repo", () => {
    const wrapper = mountWrapper(
      getStore({
        ...populatedState,
        repos: {
          reposSummaries: [{ name: "foo" } as PackageRepositorySummary],
        },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );

    // The repo name is "foo"
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "foo");
    input.simulate("change", { target: { value: "foo" } });
    // It should have pushed with the filter and fetches global repos since
    // the "kubeapps" namespace isn't the global repos namespace.
    expect(fetchRepos).toHaveBeenCalledWith("kubeapps", true);
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/kubeapps/catalog?Repository=foo"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
  });

  it("push filter for repo in other ns", () => {
    const wrapper = mountWrapper(
      getStore({
        ...populatedState,
        repos: {
          reposSummaries: [{ name: "foo" } as PackageRepositorySummary],
        },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[`/c/${defaultProps.cluster}/ns/my-ns/catalog`]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );

    // The repo name is "foo", the ns name is "my-ns"
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "foo");
    input.simulate("change", { target: { value: "foo" } });

    // It should have pushed with the filter
    expect(fetchRepos).toHaveBeenCalledWith("my-ns", true);
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/my-ns/catalog?Repository=foo"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
  });

  it("does not additionally fetch global repos when the global repo (helm plugin) is selected", () => {
    mountWrapper(
      getStore({
        ...populatedState,
        repos: { ...populatedState.repos, repos: [{ name: "foo" } as PackageRepositorySummary] },
      } as Partial<IStoreState>),
      <MemoryRouter
        initialEntries={[
          `/c/${defaultProps.cluster}/ns/${initialState.config.helmGlobalNamespace}/catalog`,
        ]}
      >
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );

    // Called without the boolean `true` option to additionally fetch global repos.
    expect(fetchRepos).toHaveBeenCalledWith("");
  });

  it("fetches from the global repos namespace for other clusters", () => {
    mountWrapper(
      getStore({
        ...populatedState,
        repos: { ...populatedState.repos, repos: [{ name: "foo" } as PackageRepositorySummary] },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[`/c/other-cluster/ns/my-ns/catalog`]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );

    // Only the global repos should have been fetched.
    expect(fetchRepos).toHaveBeenCalledWith("");
  });
});

describe("filters by operator provider", () => {
  const mockDispatch = jest.fn();

  beforeEach(() => {
    spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  });
  afterEach(() => {
    spyOnUseDispatch.mockRestore();
    mockDispatch.mockRestore();
  });

  it("doesn't show the filter if there are no csvs", () => {
    const wrapper = mountWrapper(getStore(defaultState), <Catalog />);
    expect(
      wrapper.find(FilterGroup).findWhere(g => g.prop("name") === filterNames.OPERATOR_PROVIDER),
    ).not.toExist();
  });

  const csv2 = {
    metadata: {
      name: "csv2",
    },
    spec: {
      ...csv.spec,
      provider: {
        name: "you",
      },
    },
  } as any;

  it("push filter for operator provider", () => {
    const wrapper = mountWrapper(
      getStore({
        ...populatedState,
        operators: { csvs: [csv, csv2] },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "you");
    input.simulate("change", { target: { value: "you" } });
    // It should have pushed with the filter
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/kubeapps/catalog?Provider=you"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
  });

  it("push filter for operator provider with comma", () => {
    const wrapper = mountWrapper(
      getStore({
        ...populatedState,
        operators: { csvs: [csv, csv2] },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "you");
    input.simulate("change", { target: { value: "you, inc" } });
    // It should have pushed with the filter
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/kubeapps/catalog?Provider=you__%20inc"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
  });

  it("filters by operator provider", () => {
    const wrapper = mountWrapper(
      getStore({
        ...populatedState,
        operators: { csvs: [csv, csv2] },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam + "?Provider=you"]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});

describe("filters by category", () => {
  const mockDispatch = jest.fn();

  beforeEach(() => {
    spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  });
  afterEach(() => {
    spyOnUseDispatch.mockRestore();
    mockDispatch.mockRestore();
  });
  it("renders a Unknown category if not set", () => {
    const packages = {
      ...defaultPackageState,
      items: [availablePkgSummary1],
      categories: [availablePkgSummary1.categories[0]],
    };
    const wrapper = mountWrapper(
      getStore({ ...populatedState, packages: packages } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find("input").findWhere(i => i.prop("value") === "Unknown")).toExist();
  });

  it("push filter for category", () => {
    const packages = {
      ...defaultPackageState,
      items: [availablePkgSummary1, availablePkgSummary2],
      categories: [availablePkgSummary1.categories[0], availablePkgSummary2.categories[0]],
    };
    const store = getStore({ ...defaultState, packages: packages } as IStoreState);
    const wrapper = mountWrapper(
      store,
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(2);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Database");
    input.simulate("change", { target: { value: "Database" } });
    // It should have pushed with the filter
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/kubeapps/catalog?Category=Database"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
  });

  it("filters a category", () => {
    const packages = {
      ...defaultPackageState,
      items: [availablePkgSummary1, availablePkgSummary2],
      categories: [availablePkgSummary1.categories[0], availablePkgSummary2.categories[0]],
    };
    const wrapper = mountWrapper(
      getStore({ ...populatedState, packages: packages } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam + "?Category=Database"]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("filters an operator category", () => {
    const csvWithCat = {
      ...csv,
      metadata: {
        name: "csv-cat",
        annotations: {
          categories: "E-Learning",
        },
      },
    } as any;
    const wrapper = mountWrapper(
      getStore({
        ...populatedState,
        operators: { csvs: [csv, csvWithCat] },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam + "?Category=E-Learning"]}>
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("filters operator categories", () => {
    const csvWithCat = {
      ...csv,
      metadata: {
        name: "csv-cat",
        annotations: {
          categories: "DeveloperTools, Infrastructure",
        },
      },
    } as any;
    const wrapper = mountWrapper(
      getStore({
        ...populatedState,
        operators: { csvs: [csv, csvWithCat] },
      } as Partial<IStoreState>),
      <MemoryRouter
        initialEntries={[routePathParam + "?Category=Developer%20Tools,Infrastructure"]}
      >
        <Route path={routePath}>
          <Catalog />
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});
