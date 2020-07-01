import { mount, shallow } from "enzyme";
import * as React from "react";
import Modal from "react-modal";
import * as Select from "react-select";

import { IClusterState } from "../../reducers/cluster";
import NamespaceSelector, { INamespaceSelectorProps } from "./NamespaceSelector";
import NewNamespace from "./NewNamespace";

const defaultProps = {
  fetchNamespaces: jest.fn(),
  clusters: {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "namespace-two",
        namespaces: ["namespace-one", "namespace-two"],
      } as IClusterState,
    },
  },
  defaultNamespace: "kubeapps-user",
  onChange: jest.fn(),
  createNamespace: jest.fn(),
  getNamespace: jest.fn(),
} as INamespaceSelectorProps;

it("renders the given namespaces with current selection", () => {
  const wrapper = shallow(<NamespaceSelector {...defaultProps} />);
  const select = wrapper.find(".NamespaceSelector__select").first();

  const expectedValue = {
    label: defaultProps.clusters.clusters.default.currentNamespace,
    value: defaultProps.clusters.clusters.default.currentNamespace,
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
    clusters: {
      ...defaultProps.clusters,
      clusters: {
        ...defaultProps.clusters.clusters,
        default: {
          currentNamespace: "",
          namespaces: [],
        },
      },
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
  const createNamespace = jest.fn().mockReturnValue(true);
  const wrapper = mount(<NamespaceSelector {...defaultProps} createNamespace={createNamespace} />);
  Modal.setAppElement(document.createElement("div"));
  const select = wrapper.find(Select.Creatable);
  (select.prop("onChange") as any)({ value: "_new" });
  wrapper.update();
  expect(wrapper.find(NewNamespace).prop("modalIsOpen")).toBe(true);

  wrapper.setState({ newNamespace: "test" });
  wrapper.update();

  wrapper.find(".button-primary").simulate("click");
  expect(createNamespace).toHaveBeenCalledWith("default", "test");
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
  const props = {
    ...defaultProps,
    fetchNamespaces,
    getNamespace,
    clusters: {
      ...defaultProps.clusters,
      clusters: {
        ...defaultProps.clusters.clusters,
        default: {
          currentNamespace: "foo",
          namespaces: [],
        },
      },
    },
  };
  shallow(<NamespaceSelector {...props} />);
  expect(fetchNamespaces).toHaveBeenCalled();
  expect(getNamespace).toHaveBeenCalledWith("default", "foo");
});

it("doesnt' get the current namespace if all namespaces is selected", () => {
  const fetchNamespaces = jest.fn();
  const getNamespace = jest.fn();
  const props = {
    ...defaultProps,
    fetchNamespaces,
    getNamespace,
    clusters: {
      ...defaultProps.clusters,
      clusters: {
        ...defaultProps.clusters.clusters,
        default: {
          currentNamespace: "_all",
          namespaces: [],
        },
      },
    },
  };
  shallow(<NamespaceSelector {...props} />);
  expect(fetchNamespaces).toHaveBeenCalled();
  expect(getNamespace).not.toHaveBeenCalled();
});
