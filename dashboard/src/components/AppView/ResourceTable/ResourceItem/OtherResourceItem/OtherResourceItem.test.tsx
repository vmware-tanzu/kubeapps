import { shallow } from "enzyme";
import * as React from "react";

import ResourceRef from "../../../../../shared/ResourceRef";
import { IResource } from "../../../../../shared/types";
import OtherResourceItem from "./OtherResourceItem";

it("renders a ConfigMap as a resourceref", () => {
  const cm = {
    name: "foo",
    kind: "ConfigMap",
  } as ResourceRef;
  const wrapper = shallow(<OtherResourceItem resource={cm} />);
  expect(wrapper).toMatchSnapshot();
});

it("renders a ConfigMap as a resource", () => {
  const cm = {
    metadata: {
      name: "foo",
    },
    kind: "ConfigMap",
  } as IResource;
  const wrapper = shallow(<OtherResourceItem resource={cm} />);
  expect(wrapper).toMatchSnapshot();
});
