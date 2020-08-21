import actions from "actions";
import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Modal from "components/js/Modal/Modal";
import * as React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IAppRepository } from "shared/types";
import { AppRepoAddButton } from "./AppRepoButton.v2";
import { AppRepoForm } from "./AppRepoForm.v2";

// Mocking AppRepoForm to easily test this component standalone
jest.mock("./AppRepoForm.v2", () => {
  return {
    AppRepoForm: () => <div />,
  };
});

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    updateRepo: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

const defaultProps = {
  namespace: "default",
  kubeappsNamespace: "kubeapps",
};

it("should open a modal with the repository form", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoAddButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(Modal).prop("showModal")).toBe(true);
});

it("should render a custom text", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <AppRepoAddButton {...defaultProps} text="other text" />,
  );
  expect(wrapper.find(CdsButton)).toIncludeText("other text");
});

it("should render a primary button", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoAddButton {...defaultProps} />);
  expect(wrapper.find(CdsButton).prop("action")).toBe("solid");
  expect(wrapper.find(CdsIcon)).toExist();
});

it("should render a secondary button", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <AppRepoAddButton {...defaultProps} primary={false} />,
  );
  expect(wrapper.find(CdsButton).prop("action")).toBe("outline");
  expect(wrapper.find(CdsIcon)).not.toExist();
});

it("calls installRepo when submitting", () => {
  const installRepo = jest.fn();
  actions.repos = {
    ...actions.repos,
    installRepo,
  };

  const wrapper = mountWrapper(defaultStore, <AppRepoAddButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  (wrapper.find(AppRepoForm).prop("onSubmit") as any)();
  expect(installRepo).toHaveBeenCalled();
});

it("calls updateRepo when submitting and there is a repo available", () => {
  const updateRepo = jest.fn();
  actions.repos = {
    ...actions.repos,
    updateRepo,
  };

  const wrapper = mountWrapper(
    defaultStore,
    <AppRepoAddButton
      {...defaultProps}
      repo={{ metadata: { name: "foo" }, spec: {} } as IAppRepository}
    />,
  );
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  (wrapper.find(AppRepoForm).prop("onSubmit") as any)();
  expect(updateRepo).toHaveBeenCalled();
});
