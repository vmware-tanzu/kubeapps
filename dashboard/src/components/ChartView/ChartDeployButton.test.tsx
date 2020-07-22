import * as React from "react";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IChartVersion } from "../../shared/types";
import * as url from "../../shared/url";
import ChartDeployButton, { IChartDeployButtonProps } from "./ChartDeployButton";

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
  const props = {
    version: testChartVersion,
    namespace: "kubeapps",
  } as IChartDeployButtonProps;
  const wrapper = mountWrapper(getStore({}), <ChartDeployButton {...props} />);
  const button = wrapper.find("button");
  expect(button.exists()).toBe(true);
  expect(button.text()).toBe("Deploy");
});

it("dispatches a URL change with the correct URL when the button is clicked", () => {
  const testCases = [
    {
      clustersState: { currentCluster: "default", clusters: {} },
      namespace: "kubeapps",
      version: "1.2.3",
      url: url.app.apps.new("default", "kubeapps", testChartVersion, "1.2.3"),
    },
    {
      clustersState: { currentCluster: "other-cluster", clusters: {} },
      namespace: "foo",
      version: "alpha-0",
      url: url.app.apps.new("other-cluster", "foo", testChartVersion, "alpha-0"),
    },
  ];

  testCases.forEach(t => {
    const props = {
      version: {
        ...testChartVersion,
        attributes: {
          ...testChartVersion.attributes,
          version: t.version,
        },
      },
      namespace: t.namespace,
    } as IChartDeployButtonProps;
    const store = getStore({ clusters: t.clustersState });

    const wrapper = mountWrapper(store, <ChartDeployButton {...props} />);

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
