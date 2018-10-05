import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import DeploymentItem from "./DeploymentItem";
import DeploymentTable from "./DeploymentTable";

it("renders a deployment ready", () => {
  const deployments = {};
  const dep = "foo";
  deployments[dep] = {
    kind: "Deployment",
    metadata: {
      name: "foo",
    },
    status: {},
  } as IResource;
  const wrapper = shallow(<DeploymentTable deployments={deployments} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(DeploymentItem).key()).toBe("foo");
});

it("renders two deployments", () => {
  const deployments = {};
  const dep1 = "foo";
  const dep2 = "bar";
  deployments[dep1] = { kind: "Deployment", metadata: { name: dep1 }, status: {} } as IResource;
  deployments[dep2] = { kind: "Deployment", metadata: { name: dep1 }, status: {} } as IResource;
  const wrapper = shallow(<DeploymentTable deployments={deployments} />);
  expect(wrapper.find(DeploymentItem).length).toBe(2);
  expect(
    wrapper
      .find(DeploymentItem)
      .at(0)
      .key(),
  ).toBe(dep1);
  expect(
    wrapper
      .find(DeploymentItem)
      .at(1)
      .key(),
  ).toBe(dep2);
});
