import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "../../../../../shared/types";
import DeploymentItemRow from "./DeploymentItem";

it("renders a complete Deployment", () => {
  const deployment = {
    metadata: {
      name: "foo",
    },
    status: {
      replicas: 1,
      updatedReplicas: 1,
      availableReplicas: 1,
    },
  } as IResource;
  const wrapper = shallow(<DeploymentItemRow resource={deployment} />);
  expect(wrapper).toMatchSnapshot();
});

it("completes with 0 if a status field is not populated", () => {
  const deployment = {
    metadata: {
      name: "foo",
    },
    status: {
      replicas: 1,
    },
  } as IResource;
  const wrapper = shallow(<DeploymentItemRow resource={deployment} />);
  expect(wrapper).toMatchSnapshot();
});
