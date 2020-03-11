import { shallow } from "enzyme";
import * as React from "react";
import { NotFoundError } from "../../shared/types";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import OperatorDescription from "./OperatorDescription";
import OperatorView from "./OperatorView";

const defaultProps = {
  operatorName: "foo",
  getOperator: jest.fn(),
  isFetching: false,
  namespace: "kubeapps",
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
  shallow(<OperatorView {...defaultProps} getOperator={getOperator} />);
  expect(getOperator).toHaveBeenCalledWith(defaultProps.namespace, defaultProps.operatorName);
});

it("shows an error if it exists", () => {
  const wrapper = shallow(<OperatorView {...defaultProps} error={new NotFoundError()} />);
  expect(wrapper.html()).toContain("Operator foo not found");
});

it("shows an error if the operator doesn't have any channel defined", () => {
  const operator = {
    status: {
      channels: [],
    },
  };
  const wrapper = shallow(<OperatorView {...defaultProps} operator={operator as any} />);
  expect(
    wrapper
      .find(UnexpectedErrorPage)
      .dive()
      .text(),
  ).toContain(
    "Operator foo doesn't define a valid channel. This is needed to extract required info",
  );
});

it("renders a full OperatorView", () => {
  const wrapper = shallow(<OperatorView {...defaultProps} operator={defaultOperator} />);
  expect(wrapper).toMatchSnapshot();
});

it("selects the default channel", () => {
  const operator = {
    ...defaultOperator,
    status: {
      ...defaultOperator.status,
      channels: [{ name: "alpha" }, defaultOperator.status.channels[0]],
    },
  };
  const wrapper = shallow(<OperatorView {...defaultProps} operator={operator as any} />);
  expect(wrapper.find(OperatorDescription).prop("description")).toEqual(
    "this is a testing operator",
  );
});
