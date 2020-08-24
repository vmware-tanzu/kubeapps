import actions from "actions";
import { CdsButton } from "components/Clarity/clarity";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog.v2";
import Modal from "components/js/Modal/Modal";
import * as React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IAppRepository } from "shared/types";
import { AppRepoAddButton } from "./AppRepoButton.v2";
import { AppRepoControl } from "./AppRepoControl.v2";

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    deleteRepo: jest.fn(),
    resyncRepo: jest.fn(),
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
  repo: {
    metadata: {
      name: "bitnami",
      namespace: "kubeapps",
    },
  } as IAppRepository,
};

it("deletes the repo and refreshes list", async () => {
  const deleteRepo = jest.fn();
  const fetchRepos = jest.fn();
  actions.repos = {
    ...actions.repos,
    deleteRepo,
    fetchRepos,
  };
  const wrapper = mountWrapper(defaultStore, <AppRepoControl {...defaultProps} />);
  const deleteButton = wrapper.find(CdsButton).filterWhere(b => b.text() === "Delete");
  act(() => {
    (deleteButton.prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(ConfirmDialog).find(Modal)).toIncludeText(
    "Are you sure you want to delete the repository",
  );
  const confirmButton = wrapper
    .find(ConfirmDialog)
    .find(Modal)
    .find(CdsButton)
    .filterWhere(b => b.text() === "Delete");
  await act(async () => {
    await (confirmButton.prop("onClick") as any)();
  });
  expect(deleteRepo).toHaveBeenCalled();
  expect(fetchRepos).toHaveBeenCalledWith(defaultProps.kubeappsNamespace);
});

it("deletes the repo and refreshes list (in other namespace)", async () => {
  const deleteRepo = jest.fn();
  const fetchRepos = jest.fn();
  actions.repos = {
    ...actions.repos,
    deleteRepo,
    fetchRepos,
  };
  const wrapper = mountWrapper(
    defaultStore,
    <AppRepoControl
      {...defaultProps}
      repo={
        {
          metadata: {
            name: "bitnami",
            namespace: "other",
          },
        } as IAppRepository
      }
    />,
  );
  const deleteButton = wrapper.find(CdsButton).filterWhere(b => b.text() === "Delete");
  act(() => {
    (deleteButton.prop("onClick") as any)();
  });
  wrapper.update();

  const confirmButton = wrapper
    .find(ConfirmDialog)
    .find(Modal)
    .find(CdsButton)
    .filterWhere(b => b.text() === "Delete");
  await act(async () => {
    await (confirmButton.prop("onClick") as any)();
  });
  expect(deleteRepo).toHaveBeenCalled();
  expect(fetchRepos).toHaveBeenCalledWith("other", defaultProps.kubeappsNamespace);
});

it("refreshes the repo", () => {
  const resyncRepo = jest.fn();
  actions.repos = {
    ...actions.repos,
    resyncRepo,
  };
  const wrapper = mountWrapper(defaultStore, <AppRepoControl {...defaultProps} />);
  const refreshButton = wrapper.find(CdsButton).filterWhere(b => b.text() === "Refresh");
  act(() => {
    (refreshButton.prop("onClick") as any)();
  });
  wrapper.update();
  expect(resyncRepo).toHaveBeenCalled();
});

it("renders the button to edit the repo", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoControl {...defaultProps} />);
  expect(wrapper.find(AppRepoAddButton).prop("text")).toBe("Edit");
});
