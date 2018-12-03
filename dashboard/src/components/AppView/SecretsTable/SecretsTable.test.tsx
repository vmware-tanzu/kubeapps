import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { IResource } from "shared/types";
import SecretsTable from ".";
import SecretItem from "./SecretItem";

context("before rendering", () => {
  it("should request the given names", () => {
    const getSecret = jest.fn();
    shallow(
      <SecretsTable namespace="foo" secretNames={["bar"]} secrets={{}} getSecret={getSecret} />,
    );
    expect(getSecret.mock.calls[0]).toEqual(["foo", "bar"]);
  });
});

const validProps = {
  namespace: "foo",
  secretNames: [],
  getSecret: jest.fn(),
  secrets: {},
};
it("renders a table with a secret", () => {
  const secrets = {};
  const secret = "foo";
  secrets[secret] = {
    item: {
      kind: "Secret",
      metadata: {
        name: "foo",
      },
      status: {},
    } as IResource,
  };
  const wrapper = shallow(<SecretsTable {...validProps} secrets={secrets} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(SecretItem).key()).toBe("foo");
});

it("renders a table with several secrets", () => {
  const secrets = {};
  const secret1 = "foo";
  const secret2 = "bar";
  secrets[secret1] = {
    item: { kind: "Secret", metadata: { name: secret1 }, status: {} } as IResource,
  };
  secrets[secret2] = {
    item: { kind: "Secret", metadata: { name: secret2 }, status: {} } as IResource,
  };
  const wrapper = shallow(<SecretsTable {...validProps} secrets={secrets} />);
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
