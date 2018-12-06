import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";

import SecretsTable from ".";
import LoadingWrapper from "../../../components/LoadingWrapper";
import { ISecret } from "../../../shared/types";
import SecretItem from "./SecretItem";

context("when fetching secrets", () => {
  itBehavesLike("aLoadingComponent", {
    component: SecretsTable,
    props: {
      secretNames: [],
      secrets: [{ isFetching: true }],
    },
  });
});

context("before rendering", () => {
  it("should request the given names", () => {
    const getSecret = jest.fn();
    shallow(
      <SecretsTable namespace="foo" secretNames={["bar"]} secrets={[]} getSecret={getSecret} />,
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

it("renders a message if there are no secrets", () => {
  const wrapper = shallow(<SecretsTable {...validProps} secrets={[]} />);
  expect(
    wrapper
      .find(LoadingWrapper)
      .shallow()
      .find(SecretItem),
  ).not.toExist();
  expect(
    wrapper
      .find(LoadingWrapper)
      .shallow()
      .text(),
  ).toContain("The current application does not contain any secret");
});

it("renders a table with a secret", () => {
  const secrets = [
    {
      isFetching: false,
      item: {
        kind: "Secret",
        metadata: {
          name: "foo",
        },
      } as ISecret,
    },
  ];
  const wrapper = shallow(<SecretsTable {...validProps} secrets={secrets} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(SecretItem).key()).toBe("foo");
});

it("renders a table with several secrets", () => {
  const secret1 = "foo";
  const secret2 = "bar";
  const secrets = [
    { isFetching: false, item: { kind: "Secret", metadata: { name: secret1 } } as ISecret },
    { isFetching: false, item: { kind: "Secret", metadata: { name: secret2 } } as ISecret },
  ];
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

it("renders a secret table with a secret and an error", () => {
  // const manifest = generateYamlManifest([
  //   resources.deployment,
  //   resources.service,
  //   resources.configMap,
  //   resources.ingress,
  //   resources.secret,
  // ]);
  // const wrapper = shallow(<AppViewComponent {...validProps} />);
  // validProps.app.manifest = manifest;
  // // setProps again so we trigger componentWillReceiveProps
  // wrapper.setProps(validProps);
  // const secretTable = wrapper.find(SecretTable);
  // expect(secretTable).toExist();
  // expect(secretTable.props()).toMatchObject({
  //   namespace: appRelease.namespace,
  //   secretNames: [resources.secret.metadata.name],
  // });
});
