// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog";
import { PackageRepositorySummary } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { PkgRepoAddButton } from "./PkgRepoButton";
import { IPkgRepoListItemProps, PkgRepoControl } from "./PkgRepoControl";

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    deleteRepo: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

const defaultProps = {
  helmGlobalNamespace: "kubeapps",
  carvelGlobalNamespace: "carvel-global",
  repo: {
    name: "bitnami",
    packageRepoRef: { context: { namespace: "kubeapps" } },
  } as PackageRepositorySummary,
  refetchRepos: jest.fn(),
} as IPkgRepoListItemProps;

it("deletes the repo and refreshes list", async () => {
  const deleteRepo = jest.fn();
  const refetchRepos = jest.fn();
  actions.repos = {
    ...actions.repos,
    deleteRepo,
  };
  const wrapper = mountWrapper(
    defaultStore,
    <PkgRepoControl {...defaultProps} refetchRepos={refetchRepos} />,
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

it("renders the button to edit the repo", () => {
  const wrapper = mountWrapper(defaultStore, <PkgRepoControl {...defaultProps} />);
  expect(wrapper.find(PkgRepoAddButton).prop("text")).toBe("Edit");
});
