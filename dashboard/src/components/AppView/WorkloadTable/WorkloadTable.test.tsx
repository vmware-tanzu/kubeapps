import { shallow } from "enzyme";
import * as React from "react";

import WorkloadItemContainer from "../../../containers/WorkloadItemContainer";
import ResourceRef from "../../../shared/ResourceRef";
import { IResource } from "../../../shared/types";
import WorkloadTable from "./WorkloadTable";

it("skips the element if there are no resources", () => {
  const wrapper = shallow(<WorkloadTable resourceRefs={[]} title={""} status={{}} />);
  expect(wrapper.find(WorkloadItemContainer)).not.toExist();
  expect(wrapper.html()).toBe(null);
});

it("renders a WorkloadItem", () => {
  const resourceRefs = [
    new ResourceRef({ kind: "Deployment", metadata: { name: "foo" } } as IResource, "default"),
  ];
  const wrapper = shallow(<WorkloadTable resourceRefs={resourceRefs} title={""} status={{}} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(WorkloadItemContainer)).toExist();
});

it("renders two resources", () => {
  const deployRefs = [
    new ResourceRef({ kind: "Deployment", metadata: { name: "foo" } } as IResource, "default"),
    new ResourceRef({ kind: "Deployment", metadata: { name: "bar" } } as IResource, "default"),
  ];
  const wrapper = shallow(<WorkloadTable resourceRefs={deployRefs} title={""} status={{}} />);
  expect(wrapper.find(WorkloadItemContainer).length).toBe(2);
  expect(
    wrapper
      .find(WorkloadItemContainer)
      .at(0)
      .prop("resourceRef"),
  ).toBe(deployRefs[0]);
  expect(
    wrapper
      .find(WorkloadItemContainer)
      .at(1)
      .prop("resourceRef"),
  ).toBe(deployRefs[1]);
});
