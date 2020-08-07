import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported";
import { mount, shallow } from "enzyme";
import * as React from "react";
import OperatorInstanceUpdateForm from ".";
import NotFoundErrorPage from "../ErrorAlert/NotFoundErrorAlert";
import { IOperatorInstanceUpgradeFormProps } from "./OperatorInstanceUpdateForm";

const defaultProps: IOperatorInstanceUpgradeFormProps = {
  csvName: "foo",
  crdName: "foo-cluster",
  isFetching: false,
  cluster: "default",
  namespace: "kubeapps",
  resourceName: "my-foo",
  getResource: jest.fn(),
  updateResource: jest.fn(),
  push: jest.fn(),
  errors: {},
};

const defaultResource = {
  kind: "Foo",
  apiVersion: "v1",
  metadata: {
    name: "my-foo",
  },
} as any;

it("displays an alert if rendered for an additional cluster", () => {
  const props = { ...defaultProps, cluster: "other-cluster" };
  const wrapper = shallow(<OperatorInstanceUpdateForm {...props} />);
  expect(wrapper.find(OperatorNotSupported)).toExist();
});

it("gets a resource", () => {
  const getResource = jest.fn();
  shallow(<OperatorInstanceUpdateForm {...defaultProps} getResource={getResource} />);
  expect(getResource).toHaveBeenCalledWith(
    defaultProps.namespace,
    defaultProps.csvName,
    defaultProps.crdName,
    defaultProps.resourceName,
  );
});

it("set defaultValues", () => {
  const wrapper = shallow(<OperatorInstanceUpdateForm {...defaultProps} />);
  wrapper.setProps({ resource: defaultResource });
  expect(wrapper.state()).toMatchObject({
    defaultValues: "kind: Foo\napiVersion: v1\nmetadata:\n  name: my-foo\n",
  });
});

it("renders an error if the resource is not populated", () => {
  const wrapper = shallow(<OperatorInstanceUpdateForm {...defaultProps} />);
  expect(wrapper.find(NotFoundErrorPage)).toExist();
});

it("should submit the form", () => {
  const updateResource = jest.fn();
  const wrapper = mount(
    <OperatorInstanceUpdateForm {...defaultProps} updateResource={updateResource} />,
  );
  wrapper.setProps({ resource: defaultResource });

  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(updateResource).toHaveBeenCalledWith(
    defaultProps.namespace,
    defaultResource.apiVersion,
    defaultProps.crdName,
    defaultProps.resourceName,
    defaultResource,
  );
});
