import { shallow } from "enzyme";
import * as React from "react";

import ResourceRef from "shared/ResourceRef";
import OtherResourceItem from "./OtherResourceItem";

it("renders a ConfigMap", () => {
  const cm = {
    name: "foo",
    kind: "ConfigMap",
  } as ResourceRef;
  const wrapper = shallow(<OtherResourceItem resource={cm} />);
  expect(wrapper).toMatchSnapshot();
});
