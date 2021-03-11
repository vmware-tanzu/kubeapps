import { CdsButton } from "@cds/react/button";
import actions from "actions";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IAppRepository } from "shared/types";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoControl } from "./AppRepoControl";

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
  kubeappsNamespace: "kubeapps",
  repo: {
    metadata: {
      name: "bitnami",
      namespace: "kubeapps",
    },
  } as IAppRepository,
  refetchRepos: jest.fn(),
};

it("deletes the repo and refreshes list", async () => {
  const deleteRepo = jest.fn();
  const refetchRepos = jest.fn();
  actions.repos = {
    ...actions.repos,
    deleteRepo,
  };
  const wrapper = mountWrapper(
    defaultStore,
    <AppRepoControl {...defaultProps} refetchRepos={refetchRepos} />,
  );
  const deleteButton = wrapper.find(CdsButton).filterWhere(b => b.text() === "Delete");
  act(() => {
    (deleteButton.prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(ConfirmDialog)).toIncludeText(
    "Are you sure you want to delete the repository",
  );
  const confirmButton = wrapper
    .find(ConfirmDialog)
    .find(".btn")
    .filterWhere(b => b.text() === "Delete");
  await act(async () => {
    await (confirmButton.prop("onClick") as any)();
  });
  expect(deleteRepo).toHaveBeenCalled();
  expect(refetchRepos).toHaveBeenCalled();
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
