import * as React from "react";
import * as ReactRedux from "react-redux";

import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import ClusterSelector, { IClusterSelectorProps } from "./ClusterSelector";

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
  const mockDispatch = jest.fn();
  const spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);

  const wrapper = mountWrapper(store, <ClusterSelector {...props} />);
  const select = wrapper.find("Select");
  expect(select).toExist();

  const selectOnChange = select.prop("onChange");
  expect(selectOnChange).toBeDefined();
  selectOnChange!({ value: "other" } as any);

  // fetchNamespaces returns an async thunk action - hand to test more than dispatch
  // was called once.
  expect(mockDispatch).toBeCalledTimes(1);
  expect(onChange).toBeCalledWith("other");

  spyOnUseDispatch.mockRestore();
});
