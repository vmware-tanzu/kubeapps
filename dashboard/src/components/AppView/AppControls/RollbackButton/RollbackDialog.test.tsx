import { CdsButton } from "@cds/react/button";
import { mount, shallow } from "enzyme";
import { act } from "react-dom/test-utils";
import LoadingWrapper from "../../../../components/LoadingWrapper/LoadingWrapper";
import RollbackDialog from "./RollbackDialog";

const defaultProps = {
  loading: false,
  currentRevision: 2,
  modalIsOpen: true,
  onConfirm: jest.fn(),
  closeModal: jest.fn(),
};

it("should render the loading view", () => {
  const wrapper = mount(<RollbackDialog {...defaultProps} loading={true} />);
  expect(wrapper.find(LoadingWrapper)).toHaveProp("loaded", false);
});

it("should render the form if it is not loading", () => {
  const wrapper = shallow(<RollbackDialog {...defaultProps} />);
  expect(wrapper.find("select")).toExist();
});

it("should submit the current revision", () => {
  const onConfirm = jest.fn();
  const wrapper = mount(<RollbackDialog {...defaultProps} onConfirm={onConfirm} />);
  const submit = wrapper.find(CdsButton).filterWhere(b => b.text() === "Rollback");
  expect(submit).toExist();
  act(() => {
    (submit.prop("onClick") as any)();
  });
  wrapper.update();
  expect(onConfirm).toBeCalledWith(1);
});

it("should disable the rollback button if there are no revisions", () => {
  const wrapper = mount(<RollbackDialog {...defaultProps} currentRevision={1} />);
  expect(wrapper).toIncludeText("it's not possible to rollback");
  const submit = wrapper.find(CdsButton).filterWhere(b => b.text() === "Rollback");
  expect(submit).toBeDisabled();
});
