import { mount, shallow } from "enzyme";
import * as React from "react";
import * as ReactModal from "react-modal";
import * as Select from "react-select";

import { INamespaceState } from "../../reducers/namespace";
import NamespaceSelector from "./NamespaceSelector";
import NewNamespace from "./NewNamespace";

const defaultProps = {
  fetchNamespaces: jest.fn(),
  namespace: {
    current: "namespace-two",
    namespaces: ["namespace-one", "namespace-two"],
  } as INamespaceState,
  defaultNamespace: "kubeapps-user",
  onChange: jest.fn(),
  createNamespace: jest.fn(),
};

it("renders the given namespaces with current selection", () => {
  const wrapper = shallow(<NamespaceSelector {...defaultProps} />);
  const select = wrapper.find(".NamespaceSelector__select").first();

  const expectedValue = {
    label: defaultProps.namespace.current,
    value: defaultProps.namespace.current,
  };
  expect(select.props()).toMatchObject({
    value: expectedValue,
    options: [
      { label: "All Namespaces", value: "_all" },
      { label: "namespace-one", value: "namespace-one" },
      { label: "namespace-two", value: "namespace-two" },
      { label: "Create New", value: "new" },
    ],
  });
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

  expect(select.props()).toMatchObject({
    options: [
      { label: "All Namespaces", value: "_all" },
      { label: defaultProps.defaultNamespace, value: defaultProps.defaultNamespace },
      { label: "Create New", value: "new" },
    ],
  });
});

it("opens the modal to add a new namespace and creates it", async () => {
  const createNamespace = jest.fn(() => true);
  const wrapper = mount(<NamespaceSelector {...defaultProps} createNamespace={createNamespace} />);
  ReactModal.setAppElement(document.createElement("div"));
  const select = wrapper.find(Select.Creatable);
  (select.prop("onChange") as any)({ value: "new" });
  wrapper.update();
  expect(wrapper.find(NewNamespace).prop("modalIsOpen")).toBe(true);

  wrapper.setState({ newNamespace: "test" });
  wrapper.update();

  wrapper.find(".button-primary").simulate("click");
  expect(createNamespace).toHaveBeenCalledWith("test");
  // hack to wait for the state to be updated
  await new Promise(res =>
    setTimeout(() => {
      res();
    }, 0),
  );
  wrapper.update();
  expect(wrapper.find(NewNamespace).prop("modalIsOpen")).toBe(false);
});
