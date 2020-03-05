import { shallow } from "enzyme";
import * as React from "react";
import itBehavesLike from "../../shared/specs";
import OLMNotFound from "./OLMNotFound";
import OperatorList, { IOperatorListProps } from "./OperatorList";

const defaultProps: IOperatorListProps = {
  isFetching: false,
  checkOLMInstalled: jest.fn(),
  isOLMInstalled: false,
};

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

it("render the operator list if the OLM is installed", () => {
  const wrapper = shallow(<OperatorList {...defaultProps} isOLMInstalled={true} />);
  expect(wrapper.find(OLMNotFound)).not.toExist();
});
