import { CdsButton } from "@cds/react/button";
import { CdsModal } from "@cds/react/modal";
import actions from "actions";
import Alert from "components/js/Alert";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { RollbackError } from "shared/types";
import RollbackButton from "./RollbackButton";
import ReactTooltip from "react-tooltip";

const defaultProps = {
  cluster: "default",
  namespace: "kubeapps",
  releaseName: "foo",
  revision: 3,
  releaseStatus: null,
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.apps = {
    ...actions.apps,
    rollbackApp: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("rolls back an application", async () => {
  const rollbackApp = jest.fn();
  actions.apps.rollbackApp = rollbackApp;
  const wrapper = mountWrapper(defaultStore, <RollbackButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(CdsModal)).toExist();
  wrapper
    .find("select")
    .at(0)
    .simulate("change", { target: { value: "1" } });
  await act(async () => {
    await (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Rollback")
        .prop("onClick") as any
    )();
  });
  expect(rollbackApp).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultProps.releaseName,
    1,
  );
});

it("renders an error", async () => {
  const store = getStore({ apps: { error: new RollbackError("Boom!") } });
  const wrapper = mountWrapper(store, <RollbackButton {...defaultProps} />);
  // Open modal
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();

  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});

it("should render a disabled button if when passing an in-progress status", async () => {
  const disabledProps = {
    ...defaultProps,
    releaseStatus: {
      code: 6,
    },
  };
  const wrapper = mountWrapper(defaultStore, <RollbackButton {...disabledProps} />);

  expect(wrapper.find(CdsButton)).toBeDisabled();
  expect(wrapper.find(ReactTooltip)).toExist();
});
