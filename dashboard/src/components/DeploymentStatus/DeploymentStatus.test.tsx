import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import DeploymentStatus from "./DeploymentStatus";

const defaultProps = {
  watchDeployments: jest.fn(),
  closeWatches: jest.fn(),
  deployments: [],
};

describe("componentDidMount", () => {
  it("calls watchDeployments", () => {
    const mock = jest.fn();
    shallow(<DeploymentStatus {...defaultProps} watchDeployments={mock} />);
    expect(mock).toHaveBeenCalled();
  });
});

describe("componentWillUnmount", () => {
  it("calls watchDeployments", () => {
    const mock = jest.fn();
    const wrapper = shallow(<DeploymentStatus {...defaultProps} closeWatches={mock} />);
    wrapper.unmount();
    expect(mock).toHaveBeenCalled();
  });
});

it("renders a loading status", () => {
  const deployments = [
    {
      isFetching: true,
    },
  ];
  const wrapper = shallow(<DeploymentStatus {...defaultProps} deployments={deployments} />);
  expect(wrapper.text()).toContain("Loading");
  expect(wrapper).toMatchSnapshot();
});

it("renders a deleting status", () => {
  const deployments = [
    {
      isFetching: false,
    },
  ];
  const wrapper = shallow(
    <DeploymentStatus {...defaultProps} deployments={deployments} info={{ deleted: {} }} />,
  );
  expect(wrapper.text()).toContain("Deleted");
  expect(wrapper).toMatchSnapshot();
});

it("renders a deploying status", () => {
  const deployments = [
    {
      isFetching: false,
      item: {
        status: {
          replicas: 1,
          availableReplicas: 0,
        },
      } as IResource,
    },
  ];
  const wrapper = shallow(<DeploymentStatus {...defaultProps} deployments={deployments} />);
  expect(wrapper.text()).toContain("Deploying");
  expect(wrapper).toMatchSnapshot();
});

it("renders a deployed status", () => {
  const deployments = [
    {
      isFetching: false,
      item: {
        status: {
          replicas: 1,
          availableReplicas: 1,
        },
      } as IResource,
    },
  ];
  const wrapper = shallow(<DeploymentStatus {...defaultProps} deployments={deployments} />);
  expect(wrapper.text()).toContain("Deployed");
  expect(wrapper).toMatchSnapshot();
});
