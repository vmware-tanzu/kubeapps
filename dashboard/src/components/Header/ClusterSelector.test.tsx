import * as React from "react";

import ClusterSelector, { IClusterSelectorProps } from "./ClusterSelector";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";

const defaultProps = {
  onChange: jest.fn(),
  clusters: {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "default",
        namespaces: ["default", "other"],
      },
      other: {
        currentNamespace: "default",
        namespaces: ["default", "other"],
      },
    },
  },
} as IClusterSelectorProps;

it("renders the cluster selector with the correct options", () => {
  const store = getStore({});

  const wrapper = mountWrapper(store, <ClusterSelector {...defaultProps} />);
  const options = wrapper.find("Select").prop("options");

  expect(options).toEqual([
    { label: "default", value: "default" },
    { label: "other", value: "other" },
  ]);
});

it("dispatches fetchNamespaces and calls onChange prop when changed", () => {
  const store = getStore({});
  const onChange = jest.fn();
  const props = {
    ...defaultProps,
    onChange,
  };

  const wrapper = mountWrapper(store, <ClusterSelector {...props} />);
  const select = wrapper.find("Select");
  expect(select).toExist();

  const selectOnChange = select.prop("onChange");
  expect(selectOnChange).toBeDefined();
  selectOnChange!({ value: "other" } as any);

  // TODO: how can we see the action for fetchNamespace?
  expect(store.getActions()).toEqual([]);
  expect(onChange).toBeCalledWith("other");
});
