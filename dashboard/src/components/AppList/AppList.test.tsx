import context from "jest-plugin-context";
import React from "react";
import * as ReactRedux from "react-redux";

import { deepClone } from "@cds/core/internal";
import actions from "actions";
import LoadingWrapper from "components/LoadingWrapper";
import SearchFilter from "components/SearchFilter/SearchFilter";
import * as qs from "qs";
import { act } from "react-dom/test-utils";
import * as ReactRouter from "react-router";
import { Kube } from "shared/Kube";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError, IAppOverview, IStoreState } from "../../shared/types";
import Alert from "../js/Alert";
import AppList from "./AppList";
import AppListItem from "./AppListItem";
import CustomResourceListItem from "./CustomResourceListItem";

let spyOnUseDispatch: jest.SpyInstance;
const opActions = { ...actions.operators };
const appActions = { ...actions.apps };
beforeEach(() => {
  actions.operators = {
    ...actions.operators,
    getResources: jest.fn(),
  };
  actions.apps = {
    ...actions.apps,
    fetchAppsWithUpdateInfo: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  Kube.canI = jest.fn().mockReturnValue({
    then: jest.fn(f => f(true)),
  });
});

afterEach(() => {
  actions.operators = { ...opActions };
  actions.apps = { ...appActions };
  spyOnUseDispatch.mockRestore();
});

context("when changing props", () => {
  it("should fetch apps in the new namespace", async () => {
    const fetchAppsWithUpdateInfo = jest.fn();
    const getCustomResources = jest.fn();
    actions.apps.fetchAppsWithUpdateInfo = fetchAppsWithUpdateInfo;
    actions.operators.getResources = getCustomResources;
    mountWrapper(defaultStore, <AppList />);
    expect(fetchAppsWithUpdateInfo).toHaveBeenCalledWith("default-cluster", "default");
    expect(getCustomResources).toHaveBeenCalledWith("default-cluster", "default");
  });

  it("should update the search filter", () => {
    jest.spyOn(qs, "parse").mockReturnValue({
      q: "foo",
    });
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find(SearchFilter).prop("value")).toEqual("foo");
  });

  it("should list apps in all namespaces", () => {
    jest.spyOn(qs, "parse").mockReturnValue({
      allns: "yes",
    });
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find("input[type='checkbox']")).toBeChecked();
  });

  it("should fetch apps in all namespaces", async () => {
    const fetchAppsWithUpdateInfo = jest.fn();
    const getCustomResources = jest.fn();
    actions.apps.fetchAppsWithUpdateInfo = fetchAppsWithUpdateInfo;
    actions.operators.getResources = getCustomResources;
    const wrapper = mountWrapper(defaultStore, <AppList />);
    act(() => {
      wrapper.find("input[type='checkbox']").simulate("change");
    });
    expect(fetchAppsWithUpdateInfo).toHaveBeenCalledWith("default-cluster", "");
    expect(getCustomResources).toHaveBeenCalledWith("default-cluster", "");
  });

  it("should hide the all-namespace switch if the user doesn't have permissions", async () => {
    Kube.canI = jest.fn().mockReturnValue({
      then: jest.fn((f: any) => f(false)),
    });
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find("input[type='checkbox']")).not.toExist();
  });

  describe("when store changes", () => {
    let spyOnUseState: jest.SpyInstance;
    let spyOnUseLocation: jest.SpyInstance;
    afterEach(() => {
      spyOnUseState.mockRestore();
      spyOnUseLocation.mockRestore();
    });

    it("should not set all-ns prop when getting changes in the namespace", async () => {
      const setAllNS = jest.fn();
      const useState = jest.fn();
      spyOnUseState = jest
        .spyOn(React, "useState")
        /* @ts-expect-error: Argument of type '(init: any) => any[]' is not assignable to parameter of type '() => [unknown, Dispatch<unknown>]' */
        .mockImplementation((init: any) => {
          if (init === false) {
            // Mocking the result of setAllNS
            return [false, setAllNS];
          }
          return [init, useState];
        });
      spyOnUseLocation = jest.spyOn(ReactRouter, "useLocation").mockImplementation(() => {
        return { pathname: "/foo", search: "allns=yes", state: undefined, hash: "" };
      });

      mountWrapper(defaultStore, <AppList />);
      expect(setAllNS).not.toHaveBeenCalledWith(false);
    });
  });
});

context("while fetching apps", () => {
  const state = deepClone(initialState) as IStoreState;
  state.apps.isFetching = true;

  it("behaves as loading component", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find(LoadingWrapper)).toExist();
  });

  it("behaves as loading component (for operators)", () => {
    const stateOp = deepClone(initialState) as IStoreState;
    state.operators.isFetching = true;
    const wrapper = mountWrapper(getStore(stateOp), <AppList />);
    expect(wrapper.find(LoadingWrapper)).toExist();
  });

  it("renders a Application header (while fetching)", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find("h1").text()).toContain("Applications");
  });

  it("shows the search filter and deploy button (while fetching)", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find("SearchFilter")).toExist();
    expect(wrapper.find("Link")).toExist();
  });
});

context("when fetched but not apps available", () => {
  it("renders a welcome message", () => {
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find(".applist-empty").text()).toContain("Welcome To Kubeapps");
  });

  it("shows the search filter and deploy button (no apps available)", () => {
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find("SearchFilter")).toExist();
    expect(wrapper.find("Link")).toExist();
  });
});

context("when an error is present", () => {
  const state = deepClone(initialState) as IStoreState;
  state.apps.error = new FetchError("Boom!");

  it("renders a generic error message", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(Alert).html()).toContain("Boom!");
  });

  it("renders a Application header (when error)", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find("h1").text()).toContain("Applications");
  });
});

context("when apps available", () => {
  const state = deepClone(initialState) as IStoreState;
  beforeEach(() => {
    state.apps.listOverview = [
      {
        releaseName: "foo",
        namespace: "bar",
        chartMetadata: {
          name: "bar",
          version: "1.0.0",
          appVersion: "0.1.0",
        },
        status: "deployed",
      } as IAppOverview,
    ];
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("renders a CardGrid with the available Apps", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    const itemList = wrapper.find(AppListItem);
    expect(itemList).toExist();
    expect(itemList.key()).toBe("bar/foo");
  });

  it("filters apps", () => {
    state.apps.listOverview = [
      {
        releaseName: "foo",
        namespace: "foobar",
        chartMetadata: {
          name: "foobar",
          version: "1.0.0",
          appVersion: "0.1.0",
        },
        status: "deployed",
      } as IAppOverview,
      {
        releaseName: "bar",
        namespace: "foobar",
        chartMetadata: {
          name: "foobar",
          version: "1.0.0",
          appVersion: "0.1.0",
        },
        status: "deployed",
      } as IAppOverview,
    ];
    jest.spyOn(qs, "parse").mockReturnValue({
      q: "bar",
    });
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find(AppListItem).key()).toBe("foobar/bar");
  });
});

context("when custom resources available", () => {
  const state = deepClone(initialState) as IStoreState;
  const cr = { kind: "KubeappsCluster", metadata: { name: "foo-cluster" } } as any;
  const csv = {
    metadata: {
      name: "foo",
    },
    spec: {
      customresourcedefinitions: {
        owned: [
          {
            kind: "KubeappsCluster",
          },
        ],
      },
    },
  } as any;

  beforeEach(() => {
    state.operators.resources = [cr];
    state.operators.csvs = [csv];
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("renders a CardGrid with the available resources", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).toExist();
    expect(itemList.key()).toBe("foo-cluster");
  });

  it("filters out items", () => {
    jest.spyOn(qs, "parse").mockReturnValue({
      q: "nop",
    });
    const wrapper = mountWrapper(getStore(state), <AppList />);
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).not.toExist();
  });
});
