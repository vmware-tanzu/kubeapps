import { shallow } from "enzyme";
import * as React from "react";

import { INamespaceState } from "../../reducers/namespace";
import NamespaceSelector from "./NamespaceSelector";

const defaultProps = {
  fetchNamespaces: jest.fn(),
  namespace: {
    current: "namespace-two",
    namespaces: ["namespace-one", "namespace-two"],
  } as INamespaceState,
  defaultNamespace: "kubeapps-user",
  onChange: jest.fn(),
};

it("renders the given namespaces with current selection", () => {
  const wrapper = shallow(<NamespaceSelector {...defaultProps} />);
  const select = wrapper.find(".NamespaceSelector__select").first();

  const expectedValue = {
    label: defaultProps.namespace.current,
    value: defaultProps.namespace.current,
  };
  expect(select.props()).toEqual(
    expect.objectContaining({
      value: expectedValue,
      options: [
        { label: "All Namespaces", value: "_all" },
        { label: "namespace-one", value: "namespace-one" },
        { label: "namespace-two", value: "namespace-two" },
      ],
    }),
  );
});

it("render with the default namespace selected if no current selection", () => {
  const props = {
    ...defaultProps,
    namespace: {
      ...defaultProps.namespace,
      current: "",
    },
  };
  const wrapper = shallow(<NamespaceSelector {...props} />);
  const select = wrapper.find(".NamespaceSelector__select").first();

  const expectedValue = {
    label: defaultProps.defaultNamespace,
    value: defaultProps.defaultNamespace,
  };
  expect(select.props().value).toEqual(expectedValue);
});

it("renders the default namespace option if no namespaces provided", () => {
  const props = {
    ...defaultProps,
    namespace: {
      current: "",
      namespaces: [],
    },
  };
  const wrapper = shallow(<NamespaceSelector {...props} />);
  const select = wrapper.find(".NamespaceSelector__select").first();

  expect(select.props()).toEqual(
    expect.objectContaining({
      options: [
        { label: "All Namespaces", value: "_all" },
        { label: defaultProps.defaultNamespace, value: defaultProps.defaultNamespace },
      ],
    }),
  );
});
