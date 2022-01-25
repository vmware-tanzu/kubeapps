// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import { shallow } from "enzyme";
import React from "react";
import Button from "../Button";
import Alert, { AlertIcons, AlertThemes } from "./Alert";

describe(Alert, () => {
  it("renders the required HTML structure", () => {
    const wrapper = shallow(<Alert>Test</Alert>);

    expect(wrapper).toMatchSnapshot();
  });

  it("renders the content without action and close button", () => {
    const text = "test";
    const wrapper = shallow(<Alert>{text}</Alert>);

    expect(wrapper.text()).toContain(text);
    expect(wrapper.find("button")).not.toExist();
    expect(wrapper.find(Button)).not.toExist();
  });

  it("renders a functional action button when the props are present", () => {
    const mock = jest.fn();
    const wrapper = shallow(
      <Alert action="test" onClick={mock}>
        Test
      </Alert>,
    );

    expect(wrapper.find(".alert-action")).toExist();
    wrapper.find(".alert-action").simulate("click");
    expect(mock).toHaveBeenCalled();
  });

  it("renders a functional close button when the props are present", () => {
    const mock = jest.fn();
    const wrapper = shallow(<Alert onClose={mock}>Test</Alert>);

    expect(wrapper.find(".close")).toExist();
    wrapper.find(".close").simulate("click");
    expect(mock).toHaveBeenCalled();
  });

  describe("Accessibility", () => {
    it("adds the alert role when there's no button", () => {
      const wrapper = shallow(<Alert>Test</Alert>);

      expect(wrapper.prop("role")).toBe("alert");
    });

    it("adds the alertdialog role when the notification can be closed", () => {
      const wrapper = shallow(<Alert onClose={() => {}}>Test</Alert>);

      expect(wrapper.prop("role")).toBe("alertdialog");
    });

    it("adds the alertdialog role when there's an action (App alerts)", () => {
      const wrapper = shallow(
        <Alert app action="Action" onClick={() => {}}>
          Test
        </Alert>,
      );

      expect(wrapper.prop("role")).toBe("alertdialog");
    });

    it("includes a label to notify this information is important (App alerts)", () => {
      const wrapper = shallow(<Alert app>Test</Alert>);

      expect(wrapper.prop("aria-label")).toBe("Please, read the following important information");
    });
  });

  describe("Themes", () => {
    Object.values(AlertThemes).forEach(k => {
      const theme = AlertThemes[k];
      const icon = AlertIcons[k];

      it(`apply the ${theme} theme`, () => {
        const wrapper = shallow(<Alert theme={theme}>Test</Alert>);

        expect(wrapper).toHaveClassName(`alert-${theme}`);
        expect(wrapper.find(CdsIcon).prop("shape")).toBe(icon);
      });
    });
  });

  describe("App-level Alerts", () => {
    it("adds the app-level CSS class", () => {
      const wrapper = shallow(<Alert app>Test</Alert>);

      expect(wrapper).toHaveClassName("alert-app-level");
    });

    it("adds a custom icon", () => {
      const customIcon = "objects";
      const wrapper = shallow(
        <Alert app customIcon={customIcon}>
          Test
        </Alert>,
      );

      expect(wrapper.find(CdsIcon).prop("shape")).toBe(customIcon);
    });
  });
});
