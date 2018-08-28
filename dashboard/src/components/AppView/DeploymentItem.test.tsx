import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import DeploymentItem from "./DeploymentItem";

it("renders a deployment ready", () => {
  const wrapper = shallow(
    <DeploymentItem
      deployment={
        {
          apiVersion: "extensions/v1beta1",
          kind: "Deployment",
          metadata: { name: "deployment-one" },
          status: { replicas: 1, updatedReplicas: 1, availableReplicas: 1 },
        } as IResource
      }
    />,
  );
  expect(wrapper).toMatchSnapshot();
});
