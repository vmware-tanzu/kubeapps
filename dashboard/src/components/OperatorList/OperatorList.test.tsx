import { shallow } from "enzyme";
import * as React from "react";
import itBehavesLike from "../../shared/specs";
import NotFoundErrorPage from "../ErrorAlert/NotFoundErrorAlert";
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
  expect(wrapper.find(NotFoundErrorPage)).toExist();
});

it("render the operator list if the OLM is installed", () => {
  const wrapper = shallow(<OperatorList {...defaultProps} isOLMInstalled={true} />);
  expect(wrapper.find(NotFoundErrorPage)).not.toExist();
});
