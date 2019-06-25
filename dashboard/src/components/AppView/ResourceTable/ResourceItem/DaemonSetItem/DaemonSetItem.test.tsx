import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "../../../../../shared/types";
import DaemonSetItemRow from "./DaemonSetItem";

it("renders a complete DaemonSet", () => {
  const daemonset = {
    metadata: {
      name: "foo",
    },
    status: {
      currentNumberScheduled: 1,
      numberReady: 1,
    },
  } as IResource;
  const wrapper = shallow(<DaemonSetItemRow resource={daemonset} />);
  expect(wrapper).toMatchSnapshot();
});

it("completes with 0 if a status field is not populated", () => {
  const daemonset = {
    metadata: {
      name: "foo",
    },
    status: {
      currentNumberScheduled: 1,
    },
  } as IResource;
  const wrapper = shallow(<DaemonSetItemRow resource={daemonset} />);
  expect(wrapper).toMatchSnapshot();
});
