import { shallow } from "enzyme";
import * as React from "react";

import ResourceItemContainer from "../../../containers/ResourceItemContainer";
import ResourceRef from "../../../shared/ResourceRef";
import { IResource } from "../../../shared/types";
import OtherResourceItem from "./ResourceItem/OtherResourceItem";
import ResourceTable from "./ResourceTable";

const clusterName = "cluster-name";

it("skips the element if there are no resources", () => {
  const wrapper = shallow(<ResourceTable resourceRefs={[]} title={""} />);
  expect(wrapper.find(ResourceItemContainer)).not.toExist();
  expect(wrapper.html()).toBe(null);
});

it("renders a ResourceItem", () => {
  const resourceRefs = [
    new ResourceRef(
      { kind: "Deployment", metadata: { name: "foo" } } as IResource,
      clusterName,
      "default",
    ),
  ];
  const wrapper = shallow(<ResourceTable resourceRefs={resourceRefs} title={""} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(ResourceItemContainer)).toExist();
});

it("renders two resources", () => {
  const deployRefs = [
    new ResourceRef(
      { kind: "Deployment", metadata: { name: "foo" } } as IResource,
      clusterName,
      "default",
    ),
    new ResourceRef(
      { kind: "Deployment", metadata: { name: "bar" } } as IResource,
      clusterName,
      "default",
    ),
  ];
  const wrapper = shallow(<ResourceTable resourceRefs={deployRefs} title={""} />);
  expect(wrapper.find(ResourceItemContainer).length).toBe(2);
  expect(
    wrapper
      .find(ResourceItemContainer)
      .at(0)
      .prop("resourceRef"),
  ).toBe(deployRefs[0]);
  expect(
    wrapper
      .find(ResourceItemContainer)
      .at(1)
      .prop("resourceRef"),
  ).toBe(deployRefs[1]);
});

it("renders OtherResourceItem", () => {
  const resourceRefs = [
    new ResourceRef(
      { kind: "ConfigMap", metadata: { name: "foo" } } as IResource,
      clusterName,
      "default",
    ),
  ];
  const wrapper = shallow(<ResourceTable resourceRefs={resourceRefs} title={""} />);
  expect(wrapper.find(OtherResourceItem)).toExist();
  expect(wrapper.find(ResourceItemContainer)).not.toExist();
});

it("renders OtherResource as ItemContainer", () => {
  const resourceRefs = [
    new ResourceRef(
      { kind: "ConfigMap", metadata: { name: "foo" } } as IResource,
      clusterName,
      "default",
    ),
  ];
  const wrapper = shallow(
    <ResourceTable resourceRefs={resourceRefs} title={""} requestOtherResources={true} />,
  );
  expect(wrapper.find(OtherResourceItem)).not.toExist();
  expect(wrapper.find(ResourceItemContainer)).toExist();
});
