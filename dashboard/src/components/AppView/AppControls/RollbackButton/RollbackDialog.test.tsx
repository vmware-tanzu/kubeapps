import { shallow } from "enzyme";
import * as React from "react";
import LoadingWrapper from "../../../../components/LoadingWrapper";
import RollbackDialog from "./RollbackDialog";

const defaultProps = {
  loading: false,
  currentRevision: 2,
  onConfirm: jest.fn(),
  closeModal: jest.fn(),
};

it("should render the loading view", () => {
  const wrapper = shallow(<RollbackDialog {...defaultProps} loading={true} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("should render the form if it is not loading", () => {
  const wrapper = shallow(<RollbackDialog {...defaultProps} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(LoadingWrapper)).not.toExist();
  expect(wrapper.find("select")).toExist();
});

it("should submit the current revision", () => {
  const onConfirm = jest.fn();
  const wrapper = shallow(<RollbackDialog {...defaultProps} onConfirm={onConfirm} />);
  wrapper.setState({ revision: 1 });
  const submit = wrapper.find(".button-danger");
  expect(submit).toExist();
  submit.simulate("click");
  expect(onConfirm).toBeCalledWith(1);
});
