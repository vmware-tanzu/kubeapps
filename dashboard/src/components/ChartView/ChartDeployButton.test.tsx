import { mount } from "enzyme";
import * as React from "react";
import { Provider } from "react-redux";
import configureMockStore, { MockStore } from "redux-mock-store";
import thunk from "redux-thunk";
import { IChartVersion, IStoreState } from "../../shared/types";
import * as url from "../../shared/url";
import ChartDeployButton, { IChartDeployButtonProps } from "./ChartDeployButton";

const mockStore = configureMockStore([thunk]);

// TODO(absoludity): As we move to function components with (redux) hooks we'll need to
// be including state in tests, so we may want to put things like initialState
// and a generalized getWrapper in a test helpers or similar package?
const initialState = {
  apps: {},
  auth: {},
  catalog: {},
  charts: {},
  config: {},
  kube: {},
  clusters: {},
  repos: {},
  operators: {},
} as IStoreState;

const getWrapper = (store: MockStore, props: IChartDeployButtonProps) =>
  mount(
    <Provider store={store}>
      <ChartDeployButton {...props} />
    </Provider>,
  );

const testChartVersion: IChartVersion = {
  attributes: {
    version: "1.2.3",
  },
  relationships: {
    chart: {
      data: {
        name: "test",
        repo: {
          name: "testrepo",
          namespace: "kubeapps",
        },
      },
    },
  },
} as IChartVersion;

it("renders a button to deploy the chart version", () => {
  const wrapper = getWrapper(mockStore(initialState), {
    version: testChartVersion,
    namespace: "kubeapps",
  });
  const button = wrapper.find("button");
  expect(button.exists()).toBe(true);
  expect(button.text()).toBe("Deploy");
});

it("dispatches a URL change with the correct URL when the button is clicked", () => {
  const testCases = [
    {
      clustersState: { currentCluster: "default" },
      namespace: "kubeapps",
      version: "1.2.3",
      url: url.app.apps.new(testChartVersion, "default", "kubeapps", "1.2.3"),
    },
    {
      clustersState: { currentCluster: "other-cluster" },
      namespace: "foo",
      version: "alpha-0",
      url: url.app.apps.new(testChartVersion, "other-cluster", "foo", "alpha-0"),
    },
  ];

  testCases.forEach(t => {
    const store = mockStore({ ...initialState, clusters: t.clustersState });
    const version = Object.assign({}, testChartVersion);
    version.attributes.version = t.version;

    const wrapper = getWrapper(store, { version, namespace: t.namespace });

    const button = wrapper.find("button");
    expect(button.exists()).toBe(true);
    expect(store.getActions()).toEqual([]);

    button.simulate("click");

    expect(store.getActions()).toEqual([
      {
        payload: {
          args: [t.url],
          method: "push",
        },
        type: "@@router/CALL_HISTORY_METHOD",
      },
    ]);
  });
});
