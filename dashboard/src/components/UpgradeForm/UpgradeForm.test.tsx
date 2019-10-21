import { mount, shallow } from "enzyme";
import * as React from "react";

import { IChartState, IChartVersion, UnprocessableEntity } from "../../shared/types";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import { ErrorSelector } from "../ErrorAlert";
import UpgradeForm, { IUpgradeFormProps } from "./UpgradeForm";

const defaultProps = {
  appCurrentVersion: "1.0.0",
  appCurrentValues: "foo: bar",
  chartName: "my-chart",
  namespace: "default",
  releaseName: "my-release",
  repo: "my-repo",
  selected: {} as IChartState["selected"],
  upgradeApp: jest.fn(),
  push: jest.fn(),
  goBack: jest.fn(),
  fetchChartVersions: jest.fn(),
  getChartVersion: jest.fn(),
  error: undefined,
} as IUpgradeFormProps;

const versions = [{ id: "foo", attributes: { version: "1.2.3" } }] as IChartVersion[];

describe("renders an error", () => {
  it("renders a custom error if the deployment failed", () => {
    const wrapper = shallow(
      <UpgradeForm
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
    expect(wrapper.find(ErrorSelector).exists()).toBe(true);
    expect(wrapper.find(ErrorSelector).html()).toContain(
      "Sorry! Something went wrong processing my-release",
    );
    expect(wrapper.find(ErrorSelector).html()).toContain("wrong format!");
  });
});

it("renders the full UpgradeForm", () => {
  const wrapper = shallow(
    <UpgradeForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("forwards the appValues when modified", () => {
  const wrapper = shallow(<UpgradeForm {...defaultProps} />);
  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  handleValuesChange("foo: bar");

  expect(wrapper.state("appValues")).toBe("foo: bar");
  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
});

it("forwards the valuesModifed property", () => {
  const wrapper = shallow(<UpgradeForm {...defaultProps} />);
  const handleValuesModified: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValuesModified");
  handleValuesModified("foo: bar");

  expect(wrapper.state("valuesModified")).toBe(true);
  expect(wrapper.find(DeploymentFormBody).prop("valuesModified")).toBe(true);
});

it("triggers an upgrade when submitting the form", done => {
  const releaseName = "my-release";
  const namespace = "default";
  const appValues = "foo: bar";
  const schema = { properties: { foo: { type: "string" } } };
  const upgradeApp = jest.fn(() => true);
  const push = jest.fn();
  const wrapper = mount(
    <UpgradeForm
      {...defaultProps}
      selected={{ versions, version: versions[0], schema }}
      upgradeApp={upgradeApp}
      push={push}
      namespace={namespace}
    />,
  );
  wrapper.setState({ releaseName, appValues });
  wrapper.find("form").simulate("submit");
  expect(upgradeApp).toHaveBeenCalledWith(versions[0], releaseName, namespace, appValues, schema);
  setTimeout(() => {
    expect(push).toHaveBeenCalledWith("/apps/ns/default/my-release");
    done();
  }, 1);
});
