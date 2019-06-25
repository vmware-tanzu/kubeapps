import { shallow } from "enzyme";
import * as React from "react";

import ResourceItemContainer from "../../../containers/ResourceItemContainer";
import ResourceRef from "../../../shared/ResourceRef";
import { IResource } from "../../../shared/types";
import ResourceTable from "./ResourceTable";

it("skips the element if there are no resources", () => {
  const wrapper = shallow(<ResourceTable resourceRefs={[]} title={""} />);
  expect(wrapper.find(ResourceItemContainer)).not.toExist();
  expect(wrapper.html()).toBe(null);
});

it("renders a ResourceItem", () => {
  const resourceRefs = [
    new ResourceRef({ kind: "Deployment", metadata: { name: "foo" } } as IResource, "default"),
  ];
  const wrapper = shallow(<ResourceTable resourceRefs={resourceRefs} title={""} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(ResourceItemContainer)).toExist();
});

it("renders two resources", () => {
  const deployRefs = [
    new ResourceRef({ kind: "Deployment", metadata: { name: "foo" } } as IResource, "default"),
    new ResourceRef({ kind: "Deployment", metadata: { name: "bar" } } as IResource, "default"),
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
