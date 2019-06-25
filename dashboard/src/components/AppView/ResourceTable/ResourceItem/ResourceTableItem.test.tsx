import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../../shared/specs";
import { IKubeItem, IResource } from "../../../../shared/types";
import ResourceTableItem from "./ResourceTableItem";

const kubeItem: IKubeItem<IResource> = {
  isFetching: false,
};

const defaultProps = {
  name: "foo",
  watchResource: jest.fn(),
  closeWatch: jest.fn(),
  statusFields: [],
};

describe("componentDidMount", () => {
  it("calls watchResource", () => {
    const mock = jest.fn();
    shallow(<ResourceTableItem {...defaultProps} watchResource={mock} />);
    expect(mock).toHaveBeenCalled();
  });
});

describe("componentWillUnmount", () => {
  it("calls closeWatch", () => {
    const mock = jest.fn();
    const wrapper = shallow(<ResourceTableItem {...defaultProps} closeWatch={mock} />);
    wrapper.unmount();
    expect(mock).toHaveBeenCalled();
  });
});

context("when fetching resources", () => {
  [undefined, { isFetching: true }].forEach(resource => {
    itBehavesLike("aLoadingComponent", {
      component: ResourceTableItem,
      props: {
        ...defaultProps,
        resource,
      },
    });

    it("displays the name of the resource", () => {
      const wrapper = shallow(<ResourceTableItem {...defaultProps} resource={resource} />);
      expect(wrapper.text()).toContain("foo");
    });
  });
});

context("when there is an error fetching the resource", () => {
  const resource = {
    error: new Error('deployments "foo" not found'),
    isFetching: false,
  };
  const wrapper = shallow(<ResourceTableItem {...defaultProps} resource={resource} />);

  it("diplays the resource name in the first column", () => {
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

context("when there is a valid resouce", () => {
  it("renders info about the resource status", () => {
    const deployment = {
      metadata: {
        name: "foo",
      },
      status: { replicas: 1, updatedReplicas: 1, availableReplicas: 1 },
    } as IResource;
    kubeItem.item = deployment;
    const wrapper = shallow(
      <ResourceTableItem {...defaultProps} resource={kubeItem} name={deployment.metadata.name} />,
    );
    expect(wrapper).toMatchSnapshot();
  });
});
