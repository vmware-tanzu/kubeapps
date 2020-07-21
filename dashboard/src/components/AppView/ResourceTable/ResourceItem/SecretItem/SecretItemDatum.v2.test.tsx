import { mount, shallow } from "enzyme";
import * as React from "react";

import { CdsIcon } from "components/Clarity/clarity";
import SecretItemDatum from "./SecretItemDatum.v2";

const testProps = {
  name: "foo",
  value: "YmFy", // foo
};

it("renders the secret datum (hidden by default)", () => {
  const wrapper = mount(<SecretItemDatum {...testProps} />);
  expect(wrapper.find(CdsIcon).prop("shape")).toBe("eye");
  expect(wrapper).toMatchSnapshot();
});

it("displays the secret datum value when clicking on the icon", () => {
  const wrapper = mount(<SecretItemDatum {...testProps} />);
  expect(wrapper.find(".secret-datum-text").text()).toContain("foo: 3 bytes");
  const icon = wrapper.find("button");
  icon.simulate("click");
  wrapper.update();
  expect(wrapper.find(CdsIcon).prop("shape")).toBe("eye-hide");
  expect(wrapper.find(".secret-datum-text").text()).toContain("foo: bar");
});
