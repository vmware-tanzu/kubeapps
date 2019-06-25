import { shallow } from "enzyme";
import * as React from "react";

import SecretItemDatum from "./SecretItemDatum";

const testProps = {
  name: "foo",
  value: "YmFy", // foo
};

it("renders the secret datum (hidden by default)", () => {
  const wrapper = shallow(<SecretItemDatum {...testProps} />);
  expect(wrapper.state()).toMatchObject({ hidden: true });
  expect(wrapper).toMatchSnapshot();
});

it("displays the secret datum value when clicking on the icon", () => {
  const wrapper = shallow(<SecretItemDatum {...testProps} />);
  expect(wrapper.text()).toContain("foo:3 bytes");
  const icon = wrapper.find("a");
  expect(icon).toExist();
  icon.simulate("click");
  expect(wrapper.state()).toMatchObject({ hidden: false });
  expect(wrapper.text()).toContain("foo:bar");
});
