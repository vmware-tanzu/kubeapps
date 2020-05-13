import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../../shared/specs";
import { IKubeItem, IResource } from "../../../../shared/types";
import DeploymentItemRow from "./DeploymentItem/DeploymentItem";
import OtherResourceItem from "./OtherResourceItem";
import ResourceTableItem from "./ResourceTableItem";
import SecretItem from "./SecretItem";

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
        selfLink: "/deployments/foo",
      },
      status: { replicas: 1, updatedReplicas: 1, availableReplicas: 1 },
    } as IResource;
    kubeItem.item = deployment;
    const wrapper = shallow(
      <ResourceTableItem {...defaultProps} resource={kubeItem} name={deployment.metadata.name} />,
    );
    expect(wrapper).toMatchSnapshot();
  });

  it("renders info about the resource status when given a list", () => {
    const deployment = {
      metadata: {
        name: "foo",
        selfLink: "/deployments/foo",
      },
      status: { replicas: 1, updatedReplicas: 1, availableReplicas: 1 },
    } as IResource;
    const kubeList = {
      isFetching: false,
      item: { items: [deployment] },
    };
    const wrapper = shallow(
      <ResourceTableItem
        {...defaultProps}
        resource={kubeList as any}
        name={deployment.metadata.name}
      />,
    );
    expect(wrapper.find(DeploymentItemRow)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("shows a warning when a list is empty", () => {
    const kubeList = {
      isFetching: false,
      item: { items: [] },
    };
    const wrapper = shallow(
      <ResourceTableItem {...defaultProps} resource={kubeList as any} name={""} />,
    );
    expect(wrapper.find("td")).toExist();
    expect(wrapper.find("td").text()).toEqual("No resource found");
  });

  it("skips an empty component if requested", () => {
    const kubeList = {
      isFetching: false,
      item: { items: [] },
    };
    const wrapper = shallow(
      <ResourceTableItem
        {...defaultProps}
        resource={kubeList as any}
        name={""}
        avoidEmptyResouce={true}
      />,
    );
    expect(wrapper.find("td")).not.toExist();
  });

  it("renders a ConfigMap", () => {
    const cm = {
      metadata: {
        name: "foo",
        selfLink: "/configmaps/foo",
      },
    } as IResource;
    kubeItem.item = cm;
    const wrapper = shallow(
      <ResourceTableItem {...defaultProps} resource={kubeItem} name={cm.metadata.name} />,
    );
    expect(wrapper.find(OtherResourceItem)).toExist();
  });

  it("it parses the type from the position in the self-link (it ignores the namespace)", () => {
    const deployment = {
      metadata: {
        name: "foo",
        selfLink: "/api/v1/namespaces/deployments/secrets/kubernetes",
      },
      status: { replicas: 1, updatedReplicas: 1, availableReplicas: 1 },
    } as IResource;
    kubeItem.item = deployment;
    const wrapper = shallow(
      <ResourceTableItem {...defaultProps} resource={kubeItem} name={deployment.metadata.name} />,
    );
    expect(wrapper.find(SecretItem)).toExist();
  });

  it("it parses the type from the position in the self-link (it ignores the resource name)", () => {
    const deployment = {
      metadata: {
        name: "foo",
        selfLink: "/api/v1/namespaces/default/secrets/services",
      },
      status: { replicas: 1, updatedReplicas: 1, availableReplicas: 1 },
    } as IResource;
    kubeItem.item = deployment;
    const wrapper = shallow(
      <ResourceTableItem {...defaultProps} resource={kubeItem} name={deployment.metadata.name} />,
    );
    expect(wrapper.find(SecretItem)).toExist();
  });
});
