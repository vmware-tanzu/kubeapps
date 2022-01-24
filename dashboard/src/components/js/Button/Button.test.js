// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import Button, { ButtonThemes, ButtonTypes } from ".";
import Spinner from "../Spinner";

describe(Button, () => {
  it("render the primary button as the default one", () => {
    const wrapper = shallow(<Button onClick={() => {}}>Test</Button>);
    const button = wrapper.find("button");
    expect(wrapper.prop("className")).toBe("btn button btn-primary");
    expect(button.prop("disabled")).toBe(false);
    expect(button.prop("type")).toBe("button");
    expect(button.prop("title")).toBe("");
  });

  it("render the children", () => {
    const text = "content";
    const wrapper = shallow(
      <Button onClick={() => {}}>
        <span>{text}</span>
      </Button>,
    );
    expect(wrapper.find("span")).toHaveText(text);
  });

  describe("Themes", () => {
    Object.keys(ButtonThemes).forEach(k => {
      it(`apply the ${k} theme`, () => {
        const wrapper = shallow(
          <Button theme={ButtonThemes[k]} onClick={() => {}}>
            Test
          </Button>,
        );
        expect(wrapper).toHaveClassName(`btn-${ButtonThemes[k]}`);
      });
    });
  });

  describe("Types", () => {
    Object.keys(ButtonTypes).forEach(k => {
      it(`apply the ${k} theme`, () => {
        const wrapper = shallow(
          <Button type={ButtonTypes[k]} onClick={() => {}}>
            Test
          </Button>,
        );
        expect(wrapper.find("button").prop("type")).toBe(ButtonTypes[k]);
      });
    });
  });

  describe("Outline", () => {
    it("display an outline button based on the outline prop", () => {
      const wrapper = shallow(
        <Button onClick={() => {}} outline>
          Test
        </Button>,
      );
      expect(wrapper).toHaveClassName("btn-primary-outline");
    });

    it("merge outline and theme styles", () => {
      const theme = ButtonThemes.danger;
      const wrapper = shallow(
        <Button onClick={() => {}} outline theme={theme}>
          Test
        </Button>,
      );
      expect(wrapper).toHaveClassName(`btn-${theme}-outline`);
    });
  });

  it("display a small button based on the small prop", () => {
    const wrapper = shallow(
      <Button onClick={() => {}} small>
        Test
      </Button>,
    );
    expect(wrapper).toHaveClassName("btn-sm");
  });

  it("display a flat button based on the flat prop", () => {
    const wrapper = shallow(
      <Button onClick={() => {}} flat>
        Test
      </Button>,
    );
    expect(wrapper).toHaveClassName("btn-link");
  });

  it("display a block button based on the block prop", () => {
    const wrapper = shallow(
      <Button onClick={() => {}} block>
        Test
      </Button>,
    );
    expect(wrapper).toHaveClassName("btn-block");
  });

  it("display an icon button based on the icon prop", () => {
    const wrapper = shallow(
      <Button onClick={() => {}} icon>
        <clr-icon shape="home"></clr-icon>
      </Button>,
    );
    expect(wrapper).toHaveClassName("btn-icon");
    expect(wrapper.find("clr-icon")).toExist();
  });

  it("add the title to the button", () => {
    const title = "title";
    const wrapper = shallow(
      <Button onClick={() => {}} title={title}>
        Test
      </Button>,
    );
    expect(wrapper.find("button").prop("title")).toBe(title);
  });

  it("add the deactivated status to the button", () => {
    const wrapper = shallow(
      <Button onClick={() => {}} disabled>
        Test
      </Button>,
    );
    expect(wrapper.find("button").prop("disabled")).toBe(true);
  });

  it("uses a Link when the link property is available", () => {
    const wrapper = shallow(<Button link="/">Test</Button>);
    expect(wrapper).toHaveDisplayName("Link");
  });

  it("uses an anchor when the external link property is available", () => {
    const wrapper = shallow(<Button externalLink="/">Test</Button>);
    expect(wrapper).toHaveDisplayName("a");
  });

  it("allows to add link properties when externalLink is set", () => {
    const rel = "noopener";
    const target = "_blank";
    const wrapper = shallow(
      <Button externalLink="https://example.com" target={target} rel={rel}>
        Test
      </Button>,
    );

    expect(wrapper.find("a").prop("rel")).toBe(rel);
    expect(wrapper.find("a").prop("target")).toBe(target);
  });

  it("shows a loader when the loading property is true", () => {
    const loadingText = "Downloading...";
    const wrapper = shallow(
      <Button onClick={() => {}} loading loadingText={loadingText}>
        Test
      </Button>,
    );

    expect(wrapper.find(Spinner)).toExist();
    expect(wrapper.find(Spinner).prop("text")).toBe(loadingText);
  });
});
