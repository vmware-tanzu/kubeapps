import { shallow } from "enzyme";
import * as React from "react";

import StatefulSetItem from ".";
import { IResource } from "../../../../../shared/types";

it("renders a complete DaemonSet", () => {
  const daemonset = {
    metadata: {
      name: "foo",
    },
    status: {
      replicas: 1,
      updatedReplicas: 1,
      readyReplicas: 1,
    },
  } as IResource;
  const wrapper = shallow(<StatefulSetItem resource={daemonset} />);
  expect(wrapper).toMatchSnapshot();
});

it("completes with 0 if a status field is not populated", () => {
  const daemonset = {
    metadata: {
      name: "foo",
    },
    status: {
      replicas: 1,
    },
  } as IResource;
  const wrapper = shallow(<StatefulSetItem resource={daemonset} />);
  expect(wrapper).toMatchSnapshot();
});
