import { shallow } from "enzyme";
import * as Moniker from "moniker-native";
import * as ReactRedux from "react-redux";

import ChartHeader from "components/ChartView/ChartHeader";
import Alert from "components/js/Alert";
import { act } from "react-dom/test-utils";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError, IChartState, IChartVersion } from "../../shared/types";
import * as url from "../../shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import DeploymentForm, { IDeploymentFormProps } from "./DeploymentForm";

const releaseName = "my-release";
const defaultProps = {
  chartsIsFetching: false,
  chartNamespace: "other-namespace",
  chartID: "foo",
  chartVersion: "1.0.0",
  error: undefined,
  selected: {} as IChartState["selected"],
  deployChart: jest.fn(),
  push: jest.fn(),
  fetchChartVersions: jest.fn(),
  getChartVersion: jest.fn(),
  namespace: "default",
  cluster: "default",
  kubeappsNamespace: "kubeapps",
} as IDeploymentFormProps;
const versions = [
  { id: "foo", attributes: { version: "1.2.3" }, relationships: { chart: { data: { repo: {} } } } },
] as IChartVersion[];
let monikerChooseMock: jest.Mock;

let spyOnUseDispatch: jest.SpyInstance;
beforeEach(() => {
  monikerChooseMock = jest.fn().mockReturnValue(releaseName);
  Moniker.choose = monikerChooseMock;
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  jest.resetAllMocks();
  spyOnUseDispatch.mockRestore();
});

it("fetches the available versions", () => {
  const fetchChartVersions = jest.fn();
  mountWrapper(
    defaultStore,
    <DeploymentForm {...defaultProps} fetchChartVersions={fetchChartVersions} />,
  );
  expect(fetchChartVersions).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.chartNamespace,
    defaultProps.chartID,
  );
});

describe("renders an error", () => {
  it("renders a custom error if the deployment failed", () => {
    const wrapper = shallow(
      <DeploymentForm
        {...defaultProps}
        selected={
          {
            version: { attributes: {}, relationships: { chart: { data: {} } } },
            versions: [{ id: "foo", attributes: {} }],
          } as IChartState["selected"]
        }
        error={new Error("wrong format!")}
      />,
    );
    expect(wrapper.find(Alert).exists()).toBe(true);
    expect(wrapper.find(Alert).html()).toContain("wrong format!");
  });

  it("renders a fetch error only", () => {
    const wrapper = shallow(
      <DeploymentForm
        {...defaultProps}
        selected={
          {
            version: { attributes: {}, relationships: { chart: { data: {} } } },
            versions: [{ id: "foo", attributes: {} }],
          } as IChartState["selected"]
        }
        error={new FetchError("not found")}
      />,
    );
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(ChartHeader)).not.toExist();
  });
});

it("renders a release name by default, relying in Monickers output", () => {
  monikerChooseMock.mockReturnValueOnce("foo").mockReturnValueOnce("bar");

  let wrapper = mountWrapper(
    defaultStore,
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  const name1 = wrapper.find("#releaseName").prop("value");
  expect(name1).toBe("foo");

  // When reloading the name should change
  wrapper = mountWrapper(
    defaultStore,
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  const name2 = wrapper.find("#releaseName").prop("value");
  expect(name2).toBe("bar");
});

it("forwards the appValues when modified", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  act(() => {
    handleValuesChange("foo: bar");
  });
  wrapper.update();
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
});

it("changes values if the version changes and it has not been modified", () => {
  const selected = {
    versions: [versions[0], { ...versions[0], attributes: { version: "1.2.4" } } as IChartVersion],
    version: versions[0],
    values: "bar: foo",
  };
  const wrapper = mountWrapper(
    defaultStore,
    <DeploymentForm {...defaultProps} selected={selected} />,
  );
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("bar: foo");
});

it("keep values if the version changes", () => {
  const selected = {
    versions: [versions[0], { ...versions[0], attributes: { version: "1.2.4" } } as IChartVersion],
    version: versions[0],
  };
  const wrapper = mountWrapper(
    defaultStore,
    <DeploymentForm {...defaultProps} selected={selected} />,
  );

  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  const setValuesModified: () => void = wrapper.find(DeploymentFormBody).prop("setValuesModified");
  act(() => {
    handleValuesChange("foo: bar");
    setValuesModified();
  });
  wrapper.update();
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");

  wrapper.find("select").simulate("change", { target: { value: "1.2.4" } });
  wrapper.setProps({ selected: { ...selected, values: "bar: foo" } });
  wrapper.update();
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
});

it("triggers a deployment when submitting the form", async () => {
  const namespace = "default";
  const appValues = "foo: bar";
  const schema = { properties: { foo: { type: "string", form: true } } };
  const deployChart = jest.fn().mockReturnValue(true);
  const push = jest.fn();
  const wrapper = mountWrapper(
    defaultStore,
    <DeploymentForm
      {...defaultProps}
      selected={{ versions, version: versions[0], schema }}
      deployChart={deployChart}
      push={push}
      namespace={namespace}
    />,
  );
  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  act(() => {
    handleValuesChange("foo: bar");
  });
  wrapper.update();
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");

  await act(async () => {
    // Simulating "submit" causes a console.warning
    await (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<void>)({
      preventDefault: jest.fn(),
    });
  });
  expect(deployChart).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    versions[0],
    defaultProps.chartNamespace,
    releaseName,
    appValues,
    schema,
  );
  expect(push).toHaveBeenCalledWith(url.app.apps.get("default", "default", "my-release"));
});
