import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";
import { IKubeItem, IResource } from "../../../shared/types";
import DeploymentItem from "./DeploymentItem";

const kubeItem: IKubeItem<IResource> = {
  isFetching: false,
};

describe("componentDidMount", () => {
  it("calls watchDeployment", () => {
    const mock = jest.fn();
    shallow(<DeploymentItem name="foo" watchDeployment={mock} closeWatch={jest.fn()} />);
    expect(mock).toHaveBeenCalled();
  });
});

describe("componentWillUnmount", () => {
  it("calls closeWatch", () => {
    const mock = jest.fn();
    const wrapper = shallow(
      <DeploymentItem name="foo" watchDeployment={jest.fn()} closeWatch={mock} />,
    );
    wrapper.unmount();
    expect(mock).toHaveBeenCalled();
  });
});

context("when fetching deployments", () => {
  [undefined, { isFetching: true }].forEach(deployment => {
    itBehavesLike("aLoadingComponent", {
      component: DeploymentItem,
      props: {
        deployment,
        watchDeployment: jest.fn(),
      },
    });

    it("displays the name of the deployment", () => {
      const wrapper = shallow(
        <DeploymentItem
          deployment={deployment}
          name="foo"
          watchDeployment={jest.fn()}
          closeWatch={jest.fn()}
        />,
      );
      expect(wrapper.text()).toContain("foo");
    });
  });
});

context("when there is an error fetching the Deployment", () => {
  const deployment = {
    error: new Error('deployments "foo" not found'),
    isFetching: false,
  };
  const wrapper = shallow(
    <DeploymentItem
      deployment={deployment}
      name="foo"
      watchDeployment={jest.fn()}
      closeWatch={jest.fn()}
    />,
  );

  it("diplays the Deployment name in the first column", () => {
    expect(
      wrapper
        .find("td")
        .first()
        .text(),
    ).toEqual("foo");
  });

  it("displays the error message in the second column", () => {
    expect(
      wrapper
        .find("td")
        .at(1)
        .text(),
    ).toContain('Error: deployments "foo" not found');
  });
});

context("when there is a valid Deployment", () => {
  it("renders info about the Deployment status", () => {
    const deployment = {
      metadata: {
        name: "foo",
      },
      status: { replicas: 1, updatedReplicas: 1, availableReplicas: 1 },
    } as IResource;
    kubeItem.item = deployment;
    const wrapper = shallow(
      <DeploymentItem
        deployment={kubeItem}
        name={deployment.metadata.name}
        watchDeployment={jest.fn()}
        closeWatch={jest.fn()}
      />,
    );
    expect(wrapper).toMatchSnapshot();
  });
});
