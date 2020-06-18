import { mount, shallow } from "enzyme";
import * as React from "react";
import * as ReactModal from "react-modal";
import * as Select from "react-select";

import { IClusterState } from "../../reducers/namespace";
import NamespaceSelector, { INamespaceSelectorProps } from "./NamespaceSelector";
import NewNamespace from "./NewNamespace";

const defaultProps = {
  fetchNamespaces: jest.fn(),
  cluster: {
    currentNamespace: "namespace-two",
    namespaces: ["namespace-one", "namespace-two"],
  } as IClusterState,
  defaultNamespace: "kubeapps-user",
  onChange: jest.fn(),
  createNamespace: jest.fn(),
  getNamespace: jest.fn(),
} as INamespaceSelectorProps;

it("renders the given namespaces with current selection", () => {
  const wrapper = shallow(<NamespaceSelector {...defaultProps} />);
  const select = wrapper.find(".NamespaceSelector__select").first();

  const expectedValue = {
    label: defaultProps.cluster.currentNamespace,
    value: defaultProps.cluster.currentNamespace,
  };
  expect(select.props()).toMatchObject({
    value: expectedValue,
    options: [
      { label: "All Namespaces", value: "_all" },
      { label: "namespace-one", value: "namespace-one" },
      { label: "namespace-two", value: "namespace-two" },
      { label: "Create New", value: "_new" },
    ],
  });
});

it("render with the default namespace selected if no current selection", () => {
  const props = {
    ...defaultProps,
    cluster: {
      ...defaultProps.cluster,
      currentNamespace: "",
    },
  } as INamespaceSelectorProps;
  const wrapper = shallow(<NamespaceSelector {...props} />);
  const select = wrapper.find(".NamespaceSelector__select").first();

  const expectedValue = {
    label: defaultProps.defaultNamespace,
    value: defaultProps.defaultNamespace,
  };
  expect(select.props().value).toEqual(expectedValue);
});

it("opens the modal to add a new namespace and creates it", async () => {
  const createNamespace = jest.fn(() => true);
  const wrapper = mount(<NamespaceSelector {...defaultProps} createNamespace={createNamespace} />);
  ReactModal.setAppElement(document.createElement("div"));
  const select = wrapper.find(Select.Creatable);
  (select.prop("onChange") as any)({ value: "_new" });
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

it("fetches namespaces and retrive the current namespace", () => {
  const fetchNamespaces = jest.fn();
  const getNamespace = jest.fn();
  shallow(
    <NamespaceSelector
      {...defaultProps}
      fetchNamespaces={fetchNamespaces}
      getNamespace={getNamespace}
      cluster={{ currentNamespace: "foo", namespaces: [] }}
    />,
  );
  expect(fetchNamespaces).toHaveBeenCalled();
  expect(getNamespace).toHaveBeenCalledWith("foo");
});

it("doesnt' get the current namespace if all namespaces is selected", () => {
  const fetchNamespaces = jest.fn();
  const getNamespace = jest.fn();
  shallow(
    <NamespaceSelector
      {...defaultProps}
      fetchNamespaces={fetchNamespaces}
      getNamespace={getNamespace}
      cluster={{ currentNamespace: "_all", namespaces: [] }}
    />,
  );
  expect(fetchNamespaces).toHaveBeenCalled();
  expect(getNamespace).not.toHaveBeenCalled();
});
