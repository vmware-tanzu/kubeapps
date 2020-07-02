import { mount, shallow } from "enzyme";
import * as Moniker from "moniker-native";
import * as React from "react";

import {
  ForbiddenError,
  IChartState,
  IChartVersion,
  UnprocessableEntity,
} from "../../shared/types";
import * as url from "../../shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import { PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";
import DeploymentForm, { IDeploymentFormProps } from "./DeploymentForm";

const releaseName = "my-release";
const defaultProps = {
  chartsIsFetching: false,
  kubeappsNamespace: "kubeapps",
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
} as IDeploymentFormProps;
const versions = [{ id: "foo", attributes: { version: "1.2.3" } }] as IChartVersion[];
let monikerChooseMock: jest.Mock;

beforeEach(() => {
  monikerChooseMock = jest.fn().mockReturnValue(releaseName);
  Moniker.choose = monikerChooseMock;
});

afterEach(() => {
  jest.resetAllMocks();
});

it("fetches the available versions", () => {
  const fetchChartVersions = jest.fn();
  shallow(<DeploymentForm {...defaultProps} fetchChartVersions={fetchChartVersions} />);
  expect(fetchChartVersions).toHaveBeenCalledWith(
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
            version: { attributes: {} },
            versions: [{ id: "foo", attributes: {} }],
          } as IChartState["selected"]
        }
        error={new UnprocessableEntity("wrong format!")}
      />,
    );
    wrapper.setState({ latestSubmittedReleaseName: "my-app" });
    expect(wrapper.find(UnexpectedErrorAlert).exists()).toBe(true);
    expect(wrapper.find(UnexpectedErrorAlert).html()).toContain(
      "Sorry! The installation of my-app failed",
    );
    expect(wrapper.find(UnexpectedErrorAlert).html()).toContain("wrong format!");
  });

  it("the error does not change if the release name changes", () => {
    const expectedErrorMsg = "Sorry! The installation of my-app failed";

    const wrapper = shallow(
      <DeploymentForm
        {...defaultProps}
        selected={
          {
            version: { attributes: {} },
            versions: [{ id: "foo", attributes: {} }],
          } as IChartState["selected"]
        }
        error={new UnprocessableEntity("wrong format!")}
      />,
    );

    wrapper.setState({ latestSubmittedReleaseName: "my-app" });
    expect(wrapper.find(UnexpectedErrorAlert).exists()).toBe(true);
    expect(wrapper.find(UnexpectedErrorAlert).html()).toContain(expectedErrorMsg);
    wrapper.setState({ releaseName: "another-app" });
    expect(wrapper.find(UnexpectedErrorAlert).html()).toContain(expectedErrorMsg);
  });

  it("renders a Forbidden error if the deployment fails because missing permissions", () => {
    const wrapper = mount(
      <DeploymentForm
        {...defaultProps}
        selected={
          {
            version: { attributes: {} },
            versions: [{ id: "foo", attributes: {} }],
          } as IChartState["selected"]
        }
        error={
          new ForbiddenError(
            '[{"apiGroup":"","resource":"secrets","namespace":"kubeapps","clusterWide":false,"verbs":["create","list"]}]',
          )
        }
      />,
    );
    wrapper.setState({ latestSubmittedReleaseName: "my-app" });
    expect(wrapper.find(PermissionsErrorAlert).exists()).toBe(true);
    expect(wrapper.find(PermissionsErrorAlert).text()).toContain(
      "You don't have sufficient permissions",
    );
    expect(wrapper.find(PermissionsErrorAlert).text()).toContain(
      "create, list secrets  in the kubeapps namespace",
    );
  });
});

it("renders the full DeploymentForm", () => {
  const wrapper = shallow(
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("renders a release name by default, relying in Monickers output", () => {
  monikerChooseMock.mockReturnValueOnce("foo").mockReturnValueOnce("bar");

  let wrapper = shallow(
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  const name1 = wrapper.state("releaseName") as string;
  expect(name1).toBe("foo");

  // When reloading the name should change
  wrapper = shallow(
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  const name2 = wrapper.state("releaseName") as string;
  expect(name2).toBe("bar");
});

it("forwards the appValues when modified", () => {
  const wrapper = shallow(<DeploymentForm {...defaultProps} />);
  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  handleValuesChange("foo: bar");

  expect(wrapper.state("appValues")).toBe("foo: bar");
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
});

it("forwards the valuesModifed property", () => {
  const wrapper = shallow(<DeploymentForm {...defaultProps} />);
  const handleValuesModified: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValuesModified");
  handleValuesModified("foo: bar");

  expect(wrapper.state("valuesModified")).toBe(true);
});

it("triggers a deployment when submitting the form", done => {
  const namespace = "default";
  const appValues = "foo: bar";
  const schema = { properties: { foo: { type: "string", form: true } } };
  const deployChart = jest.fn().mockReturnValue(true);
  const push = jest.fn();
  const wrapper = mount(
    <DeploymentForm
      {...defaultProps}
      selected={{ versions, version: versions[0], schema }}
      deployChart={deployChart}
      push={push}
      namespace={namespace}
    />,
  );
  wrapper.setState({ appValues });
  wrapper.find("form").simulate("submit");
  expect(deployChart).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    versions[0],
    defaultProps.chartNamespace,
    releaseName,
    appValues,
    schema,
  );
  setTimeout(() => {
    expect(push).toHaveBeenCalledWith(url.app.apps.get("default", "default", "my-release"));
    done();
  }, 1);
});
