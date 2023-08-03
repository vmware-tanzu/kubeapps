// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import AlertGroup from "components/AlertGroup";
import LoadingWrapper from "components/LoadingWrapper";
import context from "jest-plugin-context";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import ConfigLoader from ".";

it("renders a loading wrapper", () => {
  const wrapper = mountWrapper(defaultStore, <ConfigLoader loaded={false} getConfig={jest.fn()} />);
  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

context("when there is an error", () => {
  it("renders the error details", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <ConfigLoader error={new Error("Wrong config!")} getConfig={jest.fn()} />,
    );
    expect(wrapper.find(AlertGroup)).toExist();
    expect(wrapper.find(AlertGroup).text()).toContain("Wrong config!");
  });
});

describe("componentDidMount", () => {
  it("calls getConfig", () => {
    const getConfig = jest.fn();
    mountWrapper(defaultStore, <ConfigLoader getConfig={getConfig} />);
    expect(getConfig).toHaveBeenCalled();
  });
});
