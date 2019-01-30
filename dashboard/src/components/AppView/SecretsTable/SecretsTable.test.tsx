import { shallow } from "enzyme";
import * as React from "react";

import SecretsTable from ".";
import SecretItem from "./SecretItem";

it("renders a message if there are no secrets", () => {
  const wrapper = shallow(<SecretsTable secretRefs={[]} />);
  expect(wrapper.find(SecretItem)).not.toExist();
  expect(wrapper.text()).toContain("The current application does not contain any Secret objects");
});
