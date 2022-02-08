// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import { act } from "react-dom/test-utils";
import Tooltip from ".";

const defaultProps = {
  id: "test",
  label: "this is a test",
};

jest.useFakeTimers();

describe(Tooltip, () => {
  it("renders a button element", () => {
    const wrapper = shallow(<Tooltip {...defaultProps}>test</Tooltip>);

    expect(wrapper).toHaveDisplayName("button");
  });

  it("includes the content in a span inside the button", () => {
    const content = "TEST";
    const wrapper = shallow(<Tooltip {...defaultProps}>{content}</Tooltip>);

    expect(wrapper.find(".tooltip-content")).toHaveText(content);
  });

  it("closes the tooltip by default", () => {
    const wrapper = shallow(<Tooltip {...defaultProps}>test</Tooltip>);

    expect(wrapper).not.toHaveClassName("tooltip-open");
  });

  describe("open and close the tooltip", () => {
    it("based on mouse events", () => {
      const wrapper = shallow(<Tooltip {...defaultProps}>test</Tooltip>);

      expect(wrapper).not.toHaveClassName("tooltip-open");
      act(() => {
        wrapper.simulate("mouseEnter");
      });
      wrapper.update();
      expect(wrapper).toHaveClassName("tooltip-open");
      act(() => {
        wrapper.simulate("mouseLeave");
      });
      wrapper.update();
      jest.runAllTimers();
      expect(wrapper).not.toHaveClassName("tooltip-open");
    });

    it("based on focus events", () => {
      const wrapper = shallow(<Tooltip {...defaultProps}>test</Tooltip>);

      expect(wrapper).not.toHaveClassName("tooltip-open");
      act(() => {
        wrapper.simulate("focus");
      });
      wrapper.update();
      expect(wrapper).toHaveClassName("tooltip-open");
      act(() => {
        wrapper.simulate("blur");
      });
      wrapper.update();
      jest.runAllTimers();
      expect(wrapper).not.toHaveClassName("tooltip-open");
    });

    it("closes when push 'escape' key", () => {
      const wrapper = shallow(<Tooltip {...defaultProps}>test</Tooltip>);

      expect(wrapper).not.toHaveClassName("tooltip-open");
      act(() => {
        wrapper.simulate("focus");
      });
      wrapper.update();
      expect(wrapper).toHaveClassName("tooltip-open");
      act(() => {
        wrapper.simulate("keyUp", { key: "Escape" });
      });
      wrapper.update();
      expect(wrapper).not.toHaveClassName("tooltip-open");
    });
  });

  describe("Accessibility", () => {
    it("has the correct role and popup properties", () => {
      const wrapper = shallow(<Tooltip {...defaultProps}>test</Tooltip>);

      expect(wrapper.prop("role")).toBe("tooltip");
      expect(wrapper.prop("aria-haspopup")).toBe("true");
    });

    it("label and describe the information", () => {
      const wrapper = shallow(<Tooltip {...defaultProps}>test</Tooltip>);

      expect(wrapper.prop("aria-label")).toBe(defaultProps.label);
      expect(wrapper.prop("aria-describedby")).toBe(defaultProps.id);
      expect(wrapper.find("span").prop("id")).toBe(defaultProps.id);
    });

    it("set the hidden and expanded properties based on the status", () => {
      const wrapper = shallow(<Tooltip {...defaultProps}>test</Tooltip>);

      expect(wrapper.prop("aria-expanded")).toBe(false);
      expect(wrapper.find("span").prop("aria-hidden")).toBe(true);
      act(() => {
        wrapper.simulate("mouseEnter");
      });
      wrapper.update();
      expect(wrapper.prop("aria-expanded")).toBe(true);
      expect(wrapper.find("span").prop("aria-hidden")).toBe(false);
    });
  });
});
