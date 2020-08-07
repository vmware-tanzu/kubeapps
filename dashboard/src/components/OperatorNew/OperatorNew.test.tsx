import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported";
import { shallow } from "enzyme";
import * as React from "react";
import { NotFoundError } from "../../shared/types";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import OperatorNew from "./OperatorNew";

const defaultProps = {
  operatorName: "foo",
  getOperator: jest.fn(),
  isFetching: false,
  cluster: "default",
  namespace: "kubeapps",
  push: jest.fn(),
  createOperator: jest.fn(),
  errors: {},
};

const defaultOperator = {
  metadata: {
    name: "foo",
    namespace: "kubeapps",
  },
  status: {
    provider: {
      name: "Kubeapps",
    },
    defaultChannel: "beta",
    channels: [
      {
        name: "beta",
        currentCSV: "foo.1.0.0",
        currentCSVDesc: {
          displayName: "Foo",
          version: "1.0.0",
          description: "this is a testing operator",
          annotations: {
            capabilities: "Basic Install",
            repository: "github.com/kubeapps/kubeapps",
            containerImage: "kubeapps/kubeapps",
            createdAt: "one day",
          },
          installModes: [],
        },
      },
    ],
  },
} as any;

it("displays an alert if rendered for an additional cluster", () => {
  const props = { ...defaultProps, cluster: "other-cluster" };
  const wrapper = shallow(<OperatorNew {...props} />);
  expect(wrapper.find(OperatorNotSupported)).toExist();
});

it("calls getOperator when mounting the component", () => {
  const getOperator = jest.fn();
  shallow(<OperatorNew {...defaultProps} getOperator={getOperator} />);
  expect(getOperator).toHaveBeenCalledWith(defaultProps.namespace, defaultProps.operatorName);
});

it("calls getOperator when changing the namespace the component", () => {
  const getOperator = jest.fn();
  const wrapper = shallow(<OperatorNew {...defaultProps} getOperator={getOperator} />);
  expect(getOperator).toHaveBeenCalledWith(defaultProps.namespace, defaultProps.operatorName);
  wrapper.setProps({ namespace: "foo" });
  expect(getOperator).toHaveBeenCalledWith("foo", defaultProps.operatorName);
});

it("parses the default channel when receiving the operator", () => {
  const wrapper = shallow(<OperatorNew {...defaultProps} />);
  wrapper.setProps({ operator: defaultOperator });
  expect(wrapper.state()).toMatchObject({
    updateChannel: defaultOperator.status.channels[0],
    updateChannelGlobal: false,
    installationModeGlobal: false,
  });
});

it("renders an error if present", () => {
  const wrapper = shallow(
    <OperatorNew {...defaultProps} errors={{ fetch: new NotFoundError() }} />,
  );
  expect(wrapper.html()).toContain("Operator foo not found");
});

it("shows an error if the operator doesn't have any channel defined", () => {
  const operator = {
    status: {
      channels: [],
    },
  };
  const wrapper = shallow(<OperatorNew {...defaultProps} operator={operator as any} />);
  expect(
    wrapper
      .find(UnexpectedErrorPage)
      .dive()
      .text(),
  ).toContain(
    "Operator foo doesn't define a valid channel. This is needed to extract required info",
  );
});

it("renders a full OperatorNew", () => {
  const wrapper = shallow(<OperatorNew {...defaultProps} />);
  wrapper.setProps({ operator: defaultOperator });
  expect(wrapper).toMatchSnapshot();
});

it("disables the submit button if the operators ns is selected", () => {
  const wrapper = shallow(<OperatorNew {...defaultProps} namespace="operators" />);
  wrapper.setProps({ operator: defaultOperator });
  expect(wrapper.find("button").prop("disabled")).toBe(true);
  expect(wrapper.find(UnexpectedErrorPage)).toExist();
});

it("deploys an operator", async () => {
  const createOperator = jest.fn().mockReturnValue(true);
  const push = jest.fn();
  const wrapper = shallow(
    <OperatorNew
      {...defaultProps}
      namespace="operators"
      createOperator={createOperator}
      push={push}
    />,
  );
  wrapper.setProps({ operator: defaultOperator });
  const onSubmit = wrapper.find("form").prop("onSubmit") as () => Promise<void>;
  await onSubmit();

  expect(createOperator).toHaveBeenCalledWith("operators", "foo", "beta", "Automatic", "foo.1.0.0");
  expect(push).toHaveBeenCalledWith("/c/default/ns/operators/operators");
});
