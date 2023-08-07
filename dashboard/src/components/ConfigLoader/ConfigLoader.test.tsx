// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import AlertGroup from "components/AlertGroup";
import LoadingWrapper from "components/LoadingWrapper";
import context from "jest-plugin-context";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import ConfigLoader from ".";
import actions from "actions";

it("renders a loading wrapper", () => {
  const wrapper = mountWrapper(
    getStore({
      config: {
        loaded: false,
      },
    }),
    <ConfigLoader />,
  );

  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

context("when there is an error", () => {
  it("renders the error details", () => {
    const wrapper = mountWrapper(
      getStore({
        config: {
          error: new Error("Wrong config!"),
        },
      }),
      <ConfigLoader />,
    );
    expect(wrapper.find(AlertGroup)).toExist();
    expect(wrapper.find(AlertGroup).text()).toContain("Wrong config!");
  });
});

describe("componentDidMount", () => {
  it("calls getConfig", () => {
    const getConfig = jest.fn().mockReturnValue({ type: "test" });
    jest.spyOn(actions.config, "getConfig").mockImplementation(getConfig);

    mountWrapper(defaultStore, <ConfigLoader />);

    expect(getConfig).toHaveBeenCalled();
  });
});
