import { shallow } from "enzyme";
import * as React from "react";
import { IPackageManifest } from "shared/types";
import itBehavesLike from "../../shared/specs";
import { ErrorSelector } from "../ErrorAlert";
import InfoCard from "../InfoCard";
import OLMNotFound from "./OLMNotFound";
import OperatorList, { IOperatorListProps } from "./OperatorList";

const defaultProps: IOperatorListProps = {
  isFetching: false,
  checkOLMInstalled: jest.fn(),
  isOLMInstalled: false,
  operators: [],
  namespace: "default",
  getOperators: jest.fn(),
};

const sampleOperator = {
  metadata: {
    name: "foo",
  },
  status: {
    provider: {
      name: "kubeapps",
    },
    channels: [
      {
        currentCSVDesc: {
          version: "1.0.0",
          annotations: {
            categories: "security",
          },
        },
      },
    ],
  },
} as IPackageManifest;

itBehavesLike("aLoadingComponent", {
  component: OperatorList,
  props: { ...defaultProps, isFetching: true },
});

it("call the OLM check and render the NotFound message if not found", () => {
  const checkOLMInstalled = jest.fn();
  const wrapper = shallow(<OperatorList {...defaultProps} checkOLMInstalled={checkOLMInstalled} />);
  expect(checkOLMInstalled).toHaveBeenCalled();
  expect(wrapper.find(OLMNotFound)).toExist();
});

it("renders an error if exists", () => {
  const wrapper = shallow(
    <OperatorList {...defaultProps} isOLMInstalled={true} error={new Error("Boom!")} />,
  );
  expect(wrapper.find(ErrorSelector)).toExist();
  expect(
    wrapper
      .find(ErrorSelector)
      .dive()
      .dive()
      .text(),
  ).toMatch("Boom!");
});

it("skips the error if the OLM is not installed", () => {
  const wrapper = shallow(
    <OperatorList
      {...defaultProps}
      isOLMInstalled={false}
      error={new Error("There are no operators!")}
    />,
  );
  expect(wrapper.find(ErrorSelector)).not.toExist();
  expect(wrapper.find(OLMNotFound)).toExist();
});

it("render the operator list if the OLM is installed", () => {
  const wrapper = shallow(
    <OperatorList {...defaultProps} isOLMInstalled={true} operators={[sampleOperator]} />,
  );
  expect(wrapper.find(OLMNotFound)).not.toExist();
  expect(wrapper.find(InfoCard)).toExist();
  expect(wrapper).toMatchSnapshot();
});
