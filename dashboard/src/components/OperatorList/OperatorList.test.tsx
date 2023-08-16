// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import "@testing-library/jest-dom";
import { act, screen } from "@testing-library/react";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import SearchFilter from "components/SearchFilter/SearchFilter";
import * as ReactRedux from "react-redux";
import { Route, Routes } from "react-router-dom";
import {
  defaultStore,
  getStore,
  initialState,
  mountWrapper,
  renderWithProviders,
} from "shared/specs/mountWrapper";
import { IPackageManifest, IStoreState } from "shared/types";
import InfoCard from "../InfoCard/InfoCard";
import { AUTO_PILOT, BASIC_INSTALL } from "../OperatorView/OperatorCapabilityLevel";
import OLMNotFound from "./OLMNotFound";
import OperatorItems from "./OperatorItems";
import OperatorList, { filterNames } from "./OperatorList";

let spyOnUseDispatch: jest.SpyInstance;
const kubeActions = { ...actions.operators };
beforeEach(() => {
  actions.operators = {
    ...actions.operators,
    checkOLMInstalled: jest.fn(),
    getOperators: jest.fn(),
    getCSVs: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
});

const sampleOperator = {
  metadata: {
    name: "foo",
  },
  status: {
    provider: {
      name: "kubeapps",
    },
    defaultChannel: "alpha",
    channels: [
      {
        name: "alpha",
        currentCSV: "kubeapps-operator",
        currentCSVDesc: {
          version: "1.0.0",
          annotations: {
            categories: "security",
            capabilities: AUTO_PILOT,
          },
        },
      },
    ],
  },
} as IPackageManifest;

const sampleSubscription = {
  metadata: { name: "kubeapps-operator" },
  spec: { name: "foo" },
} as any;

it("renders a LoadingWrapper if fetching", () => {
  const wrapper = mountWrapper(
    getStore({
      ...initialState,
      operators: { ...initialState.operators, isFetching: true },
    } as Partial<IStoreState>),
    <OperatorList />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("call the OLM check and render the NotFound message if not found", () => {
  const checkOLMInstalled = jest.fn();
  actions.operators.checkOLMInstalled = checkOLMInstalled;
  const wrapper = mountWrapper(defaultStore, <OperatorList />);
  expect(checkOLMInstalled).toHaveBeenCalled();
  expect(wrapper.find(OLMNotFound)).toExist();
});

it("renders an error", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { isOLMInstalled: true, errors: { operator: { fetch: new Error("Forbidden!") } } },
    } as Partial<IStoreState>),
    <OperatorList />,
  );
  const error = wrapper.find(AlertGroup).filterWhere(a => a.prop("status") === "danger");
  expect(error).toExist();
  expect(error).toIncludeText("Forbidden!");
});

it("request operators if the OLM is installed", () => {
  const getOperators = jest.fn();
  actions.operators.getOperators = getOperators;
  const wrapper = mountWrapper(
    getStore({ operators: { isOLMInstalled: true } } as Partial<IStoreState>),
    <OperatorList />,
  );
  wrapper.setProps({ namespace: "other" });
  expect(getOperators).toHaveBeenCalled();
});

it("render the operator list", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { isOLMInstalled: true, operators: [sampleOperator] },
    } as Partial<IStoreState>),
    <OperatorList />,
  );
  expect(wrapper.find(OLMNotFound)).not.toExist();
  expect(wrapper.find(InfoCard)).toExist();
});

it("render the operator list with installed operators", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: {
        isOLMInstalled: true,
        operators: [sampleOperator],
        subscriptions: [sampleSubscription],
      },
    } as Partial<IStoreState>),
    <OperatorList />,
  );
  expect(wrapper.find(OLMNotFound)).not.toExist();
  expect(wrapper.find(InfoCard)).toExist();
  // The section "Available operators" should be empty since all the ops are installed
  expect(wrapper.find("h3").filterWhere(c => c.text() === "Installed")).toExist();
  const operatorLists = wrapper.find(OperatorItems);
  expect(operatorLists).toHaveLength(2);
  expect(operatorLists.at(0)).toHaveProp("operators", [sampleOperator]);
  expect(operatorLists.at(1)).toHaveProp("operators", []);
});

it("render the operator list without installed operators", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { isOLMInstalled: true, operators: [sampleOperator] },
    } as Partial<IStoreState>),
    <OperatorList />,
  );
  expect(wrapper.find(OLMNotFound)).not.toExist();
  expect(wrapper.find(InfoCard)).toExist();
  // The section "Available operators" should not be empty since the operator is not installed
  expect(wrapper.find("h3").filterWhere(c => c.text() === "Installed")).not.toExist();
  const operatorLists = wrapper.find(OperatorItems);
  expect(operatorLists).toHaveLength(1);
  expect(operatorLists.at(0)).toHaveProp("operators", [sampleOperator]);
});

describe("filter operators", () => {
  const sampleOperator2 = {
    ...sampleOperator,
    metadata: {
      name: "bar",
    },
    status: {
      ...sampleOperator.status,
      channels: [
        {
          ...sampleOperator.status.channels[0],
          currentCSVDesc: {
            version: "1.0.0",
            annotations: {
              categories: "database, other",
              capabilities: BASIC_INSTALL,
            },
          },
        },
      ],
    },
  } as any;

  it("setting the filter in the state", () => {
    const wrapper = mountWrapper(
      getStore({
        operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
      } as Partial<IStoreState>),
      <OperatorList />,
    );
    expect(wrapper.find(InfoCard).length).toBe(2);
    act(() => {
      (wrapper.find(SearchFilter).prop("onChange") as any)("foo");
    });
    wrapper.update();
    const operator = wrapper.find(InfoCard);
    expect(operator.length).toBe(1);
    expect(operator.prop("title")).toBe(sampleOperator.metadata.name);
  });

  it("setting the filter in the query string", () => {
    const store = getStore({
      operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
    } as Partial<IStoreState>);
    renderWithProviders(
      <Routes>
        <Route path={"/c/:cluster/ns/:namespace/operators"} element={<OperatorList />} />
      </Routes>,
      {
        store,
        initialEntries: [
          `/c/default/ns/default/operators?${filterNames.SEARCH}=${sampleOperator.metadata.name}`,
        ],
      },
    );

    expect(screen.getByText(`Search: ${sampleOperator.metadata.name}`)).toBeInTheDocument();
  });

  it("transforms the received '__' in query params into a ','", () => {
    const store = getStore({
      operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
    } as Partial<IStoreState>);
    renderWithProviders(
      <Routes>
        <Route path={"/c/:cluster/ns/:namespace/operators"} element={<OperatorList />} />
      </Routes>,
      {
        store,
        initialEntries: [
          `/c/default/ns/default/operators?${filterNames.PROVIDER}=kubeapps__%20inc`,
        ],
      },
    );

    expect(screen.getByText("Provider: kubeapps, inc")).toBeInTheDocument();
  });

  it("show a message if the filter doesn't match any operator", () => {
    const store = getStore({
      operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
    } as Partial<IStoreState>);
    renderWithProviders(
      <Routes>
        <Route path={"/c/:cluster/ns/:namespace/operators"} element={<OperatorList />} />
      </Routes>,
      {
        store,
        initialEntries: [`/c/default/ns/default/operators?${filterNames.SEARCH}=nope`],
      },
    );

    expect(screen.getByText("No operator matches the current filter.")).toBeInTheDocument();
  });

  it("filters by category", () => {
    const store = getStore({
      operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
    } as Partial<IStoreState>);
    renderWithProviders(
      <Routes>
        <Route path={"/c/:cluster/ns/:namespace/operators"} element={<OperatorList />} />
      </Routes>,
      {
        store,
        initialEntries: [`/c/default/ns/default/operators?${filterNames.CATEGORY}=security`],
      },
    );

    expect(screen.getByText(sampleOperator.metadata.name)).toHaveAttribute("class", "card-title");
  });

  it("filters by capability", () => {
    const store = getStore({
      operators: { isOLMInstalled: true, operators: [sampleOperator, sampleOperator2] },
    } as Partial<IStoreState>);
    renderWithProviders(
      <Routes>
        <Route path={"/c/:cluster/ns/:namespace/operators"} element={<OperatorList />} />
      </Routes>,
      {
        store,
        initialEntries: [
          `/c/default/ns/default/operators?${filterNames.CAPABILITY}=${BASIC_INSTALL}`,
        ],
      },
    );

    expect(screen.getByText(sampleOperator2.metadata.name)).toHaveAttribute("class", "card-title");
  });
});
