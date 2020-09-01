import { mount } from "enzyme";
import * as React from "react";
import { Provider } from "react-redux";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import { CdsButton } from "components/Clarity/clarity";
import { cloneDeep } from "lodash";
import { act } from "react-dom/test-utils";
import { IClustersState } from "reducers/cluster";
import ContextSelector from "./ContextSelector";

const mockStore = configureMockStore([thunk]);
const defaultStore = mockStore({});

const defaultProps = {
  fetchNamespaces: jest.fn(),
  clusters: {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "default",
        namespaces: ["default", "other"],
      },
    },
  } as IClustersState,
  defaultNamespace: "kubeapps-user",
  setNamespace: jest.fn(),
  createNamespace: jest.fn(),
  getNamespace: jest.fn(),
};

it("fetches namespaces", () => {
  const fetchNamespaces = jest.fn();
  const getNamespace = jest.fn();
  mount(
    <Provider store={defaultStore}>
      <ContextSelector
        {...defaultProps}
        fetchNamespaces={fetchNamespaces}
        getNamespace={getNamespace}
      />
      ,
    </Provider>,
  );

  expect(fetchNamespaces).toHaveBeenCalled();
  expect(getNamespace).toHaveBeenCalledWith(
    defaultProps.clusters.currentCluster,
    defaultProps.clusters.clusters.default.currentNamespace,
  );
});

it("opens the dropdown menu", () => {
  const wrapper = mount(
    <Provider store={defaultStore}>
      <ContextSelector {...defaultProps} />
    </Provider>,
  );
  expect(wrapper.find(".dropdown")).not.toHaveClassName("open");
  const menu = wrapper.find("button");
  menu.simulate("click");
  wrapper.update();
  expect(wrapper.find(".dropdown")).toHaveClassName("open");
});

it("selects a different namespace", () => {
  const setNamespace = jest.fn();
  const wrapper = mount(
    <Provider store={defaultStore}>
      <ContextSelector {...defaultProps} setNamespace={setNamespace} />
    </Provider>,
  );
  wrapper
    .find("select")
    .findWhere(s => s.prop("name") === "namespaces")
    .simulate("change", { target: { value: "other" } });
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  expect(setNamespace).toHaveBeenCalledWith("other");
});

it("shows the current cluster", () => {
  const clusters = {
    currentCluster: "bar",
    clusters: {
      foo: {
        currentNamespace: "default",
        namespaces: ["default"],
      },
      bar: {
        currentNamespace: "default",
        namespaces: ["default"],
      },
    },
  } as IClustersState;
  const wrapper = mount(
    <Provider store={defaultStore}>
      <ContextSelector {...defaultProps} clusters={clusters} />
    </Provider>,
  );
  expect(
    wrapper
      .find("select")
      .at(0)
      .prop("value"),
  ).toBe("bar");
});

it("shows the current namespace", () => {
  const props = cloneDeep(defaultProps);
  props.clusters.clusters.default.currentNamespace = "other";
  const wrapper = mount(
    <Provider store={defaultStore}>
      <ContextSelector {...props} />
    </Provider>,
  );
  expect(
    wrapper
      .find("select")
      .at(1)
      .prop("value"),
  ).toBe("other");
});
