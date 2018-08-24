import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import DeploymentTable from "./DeploymentTable";
import DeploymentItem from "./DeploymentItem";

it("renders a deployment ready", () => {
  const deployments = new Map<string, IResource>();
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
