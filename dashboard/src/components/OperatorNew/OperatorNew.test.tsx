import { shallow } from "enzyme";
import * as React from "react";
import { NotFoundError } from "../../shared/types";
import { wait } from "../../shared/utils";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import OperatorNew from "./OperatorNew";

const defaultProps = {
  operatorName: "foo",
  getOperator: jest.fn(),
  isFetching: false,
  namespace: "kubeapps",
  push: jest.fn(),
  createOperator: jest.fn(),
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

it("shows an error if it exists", () => {
  const wrapper = shallow(<OperatorNew {...defaultProps} error={new NotFoundError()} />);
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
  const createOperator = jest.fn(() => true);
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
  wrapper.find("form").simulate("submit");
  await wait(1);

  expect(createOperator).toHaveBeenCalledWith("operators", "foo", "beta", "Automatic", "foo.1.0.0");
  expect(push).toHaveBeenCalledWith("/ns/operators/operators");
});
