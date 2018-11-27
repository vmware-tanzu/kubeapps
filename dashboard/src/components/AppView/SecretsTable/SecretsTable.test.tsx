import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import SecretsTable from ".";
import SecretItem from "./SecretItem";

it("renders a table with a secret", () => {
  const secrets = {};
  const secret = "foo";
  secrets[secret] = {
    kind: "Secret",
    metadata: {
      name: "foo",
    },
    status: {},
  } as IResource;
  const wrapper = shallow(<SecretsTable secrets={secrets} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(SecretItem).key()).toBe("foo");
});

it("renders a table with several secrets", () => {
  const secrets = {};
  const secret1 = "foo";
  const secret2 = "bar";
  secrets[secret1] = { kind: "Secret", metadata: { name: secret1 }, status: {} } as IResource;
  secrets[secret2] = { kind: "Secret", metadata: { name: secret2 }, status: {} } as IResource;
  const wrapper = shallow(<SecretsTable secrets={secrets} />);
  expect(wrapper.find(SecretItem).length).toBe(2);
  expect(
    wrapper
      .find(SecretItem)
      .at(0)
      .key(),
  ).toBe(secret1);
  expect(
    wrapper
      .find(SecretItem)
      .at(1)
      .key(),
  ).toBe(secret2);
});
