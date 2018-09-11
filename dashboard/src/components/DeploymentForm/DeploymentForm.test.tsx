import { shallow } from "enzyme";
import * as React from "react";
import { IChartState, IChartVersion, NotFoundError } from "../../shared/types";
import NotFoundErrorPage from "../ErrorAlert/NotFoundErrorAlert";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import DeploymentForm from "./DeploymentForm";

const defaultProps = {
  kubeappsNamespace: "kubeapps",
  bindingsWithSecrets: [],
  chartID: "foo",
  chartVersion: "1.0.0",
  error: undefined,
  selected: {} as IChartState["selected"],
  deployChart: jest.fn(),
  push: jest.fn(),
  fetchChartVersions: jest.fn(),
  getBindings: jest.fn(),
  getChartVersion: jest.fn(),
  getChartValues: jest.fn(),
  namespace: "default",
};

it("renders a loading message if the selected chart is not ready", () => {
  const wrapper = shallow(<DeploymentForm {...defaultProps} />);
  expect(wrapper.text()).toContain("Loading");
});

describe("renders an error", () => {
  it("renders an error if it cannot find the given chart", () => {
    const wrapper = shallow(
      <DeploymentForm
        {...defaultProps}
        selected={{ error: new NotFoundError() } as IChartState["selected"]}
      />,
    );
    expect(wrapper.find(NotFoundErrorPage).exists()).toBe(true);
  });

  it("renders a generic error", () => {
    const wrapper = shallow(
      <DeploymentForm
        {...defaultProps}
        selected={{ error: new Error() } as IChartState["selected"]}
      />,
    );
    expect(wrapper.find(UnexpectedErrorPage).exists()).toBe(true);
  });
});

it("renders the full DeploymentForm", () => {
  const versions = [{ id: "foo", attributes: { version: "1.2.3" } }] as IChartVersion[];
  const wrapper = shallow(
    <DeploymentForm {...defaultProps} selected={{ versions, version: versions[0] }} />,
  );
  expect(wrapper).toMatchSnapshot();
});
