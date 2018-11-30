import { shallow } from "enzyme";
import * as React from "react";

import { ISecret } from "shared/types";
import SecretsTable from ".";
import SecretItem from "./SecretItem";

const secret = {
  kind: "Secret",
  apiVersion: "v1",
  type: "Opaque",
  metadata: {
    name: "foo",
  },
  data: {},
} as ISecret;

it("renders a table with a secret", () => {
  const wrapper = shallow(<SecretsTable secrets={[secret]} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(SecretItem).key()).toBe("foo");
});

it("renders a table with several secrets", () => {
  const secret2 = Object.assign({}, secret);
  secret2.metadata.name = "bar";
  const wrapper = shallow(<SecretsTable secrets={[secret, secret2]} />);
  expect(wrapper.find(SecretItem).length).toBe(2);
  expect(
    wrapper
      .find(SecretItem)
      .at(0)
      .key(),
  ).toBe(secret);
  expect(
    wrapper
      .find(SecretItem)
      .at(1)
      .key(),
  ).toBe(secret2);
});
