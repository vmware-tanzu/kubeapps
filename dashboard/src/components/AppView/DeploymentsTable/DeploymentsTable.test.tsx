import { shallow } from "enzyme";
import * as React from "react";

import DeploymentItemContainer from "../../../containers/DeploymentItemContainer";
import ResourceRef from "../../../shared/ResourceRef";
import { IResource } from "../../../shared/types";
import DeploymentTable from "./DeploymentsTable";

it("renders a message if there are no deployments", () => {
  const wrapper = shallow(<DeploymentTable deployRefs={[]} />);
  expect(wrapper.find(DeploymentItemContainer)).not.toExist();
  expect(wrapper.text()).toContain(
    "The current application does not contain any Deployment objects",
  );
});

it("renders a DeploymentItem", () => {
  const deployRefs = [
    new ResourceRef({ kind: "Deployment", metadata: { name: "foo" } } as IResource, "default"),
  ];
  const wrapper = shallow(<DeploymentTable deployRefs={deployRefs} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(DeploymentItemContainer)).toExist();
});

it("renders two deployments", () => {
  const deployRefs = [
    new ResourceRef({ kind: "Deployment", metadata: { name: "foo" } } as IResource, "default"),
    new ResourceRef({ kind: "Deployment", metadata: { name: "bar" } } as IResource, "default"),
  ];
  const wrapper = shallow(<DeploymentTable deployRefs={deployRefs} />);
  expect(wrapper.find(DeploymentItemContainer).length).toBe(2);
  expect(
    wrapper
      .find(DeploymentItemContainer)
      .at(0)
      .prop("deployRef"),
  ).toBe(deployRefs[0]);
  expect(
    wrapper
      .find(DeploymentItemContainer)
      .at(1)
      .prop("deployRef"),
  ).toBe(deployRefs[1]);
});
