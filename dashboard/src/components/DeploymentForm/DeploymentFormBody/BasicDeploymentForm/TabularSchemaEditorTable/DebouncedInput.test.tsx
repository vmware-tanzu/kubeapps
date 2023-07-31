// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import { mount } from "enzyme";
import DebouncedInput from "./DebouncedInput";

jest.useFakeTimers();

it("should debounce a change in the input value", () => {
  const onChange = jest.fn();

  const wrapper = mount(<DebouncedInput value={"initial"} onChange={onChange} />);

  act(() => {
    (wrapper.find("input").prop("onChange") as any)({ target: { value: "something" } });
  });
  wrapper.update();

  expect(onChange).not.toHaveBeenCalled();
  jest.runAllTimers();
  expect(onChange).toHaveBeenCalledWith("something");
});
