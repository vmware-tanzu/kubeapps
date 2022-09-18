// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsModal } from "@cds/react/modal";
import actions from "actions";
import { PackageRepositoryReference } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IPkgRepoAddButtonProps, PkgRepoAddButton } from "./PkgRepoButton";
import { PkgRepoForm } from "./PkgRepoForm";

// Mocking PkgRepoForm to easily test this component standalone
/* eslint-disable react/display-name */
jest.mock("./PkgRepoForm", () => {
  return {
    PkgRepoForm: () => <div />,
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
  helmGlobalNamespace: "kubeapps",
  carvelGlobalNamespace: "carvel-global",
} as IPkgRepoAddButtonProps;

it("should open a modal with the repository form", () => {
  const wrapper = mountWrapper(defaultStore, <PkgRepoAddButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(CdsModal)).toExist();
});

it("should render a custom text", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PkgRepoAddButton {...defaultProps} text="other text" />,
  );
  expect(wrapper.find(CdsButton)).toIncludeText("other text");
});

it("should render a primary button", () => {
  const wrapper = mountWrapper(defaultStore, <PkgRepoAddButton {...defaultProps} />);
  expect(wrapper.find(CdsButton).prop("action")).toBe("solid");
  expect(wrapper.find(CdsIcon)).toExist();
});

it("should render a secondary button", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PkgRepoAddButton {...defaultProps} primary={false} />,
  );
  expect(wrapper.find(CdsButton).prop("action")).toBe("outline");
  expect(wrapper.find(CdsIcon)).not.toExist();
});

it("calls addRepo when submitting", () => {
  const addRepo = jest.fn();
  actions.repos = {
    ...actions.repos,
    addRepo,
  };

  const wrapper = mountWrapper(defaultStore, <PkgRepoAddButton {...defaultProps} />);
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  (wrapper.find(PkgRepoForm).prop("onSubmit") as any)();
  expect(addRepo).toHaveBeenCalled();
});

it("calls updateRepo when submitting and there is a repo available", () => {
  const updateRepo = jest.fn();
  actions.repos = {
    ...actions.repos,
    updateRepo,
  };

  const wrapper = mountWrapper(
    defaultStore,
    <PkgRepoAddButton
      {...defaultProps}
      packageRepoRef={
        {
          identifier: "foo",
        } as PackageRepositoryReference
      }
    />,
  );
  act(() => {
    (wrapper.find(CdsButton).prop("onClick") as any)();
  });
  wrapper.update();
  (wrapper.find(PkgRepoForm).prop("onSubmit") as any)();
  expect(updateRepo).toHaveBeenCalled();
});

it("should deactivate the button if given", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PkgRepoAddButton {...defaultProps} disabled={true} />,
  );
  expect(wrapper.find(CdsButton)).toBeDisabled();
});

it("should use the given title", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <PkgRepoAddButton {...defaultProps} title={"a title"} />,
  );
  expect(wrapper.find(CdsButton).prop("title")).toBe("a title");
});
