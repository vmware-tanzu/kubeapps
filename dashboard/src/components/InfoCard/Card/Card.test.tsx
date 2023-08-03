// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount, shallow } from "enzyme";
import React from "react";
import Card, { CardBlock, CardFooter, CardHeader, CardText, CardTitle } from ".";

describe(Card, () => {
  it("renders a simple card", () => {
    const wrapper = shallow(
      <Card>
        <CardBlock>My Card</CardBlock>
      </Card>,
    );
    expect(wrapper).toMatchSnapshot();
  });

  it("renders inner card elements", () => {
    const header = "Header";
    const body = "Text";
    const footer = "Footer";
    const wrapper = mount(
      <Card>
        <CardHeader>
          <CardTitle level={1}>{header}</CardTitle>
        </CardHeader>
        <CardBlock>
          <CardText>{body}</CardText>
        </CardBlock>
        <CardFooter>
          <CardText>{footer}</CardText>
        </CardFooter>
      </Card>,
    );
    expect(wrapper.find(CardHeader)).toExist();
    expect(wrapper.find(CardTitle)).toExist();
    expect(wrapper.find(CardBlock)).toExist();
    expect(wrapper.find(CardFooter)).toExist();
    expect(wrapper.find(CardText)).toExist();
    expect(wrapper).toIncludeText(header);
    expect(wrapper).toIncludeText(body);
    expect(wrapper).toIncludeText(footer);
  });

  it("set the HTML tag based on a prop", () => {
    const wrapper = shallow(
      <Card htmlTag="article">
        <CardBlock>My Card</CardBlock>
      </Card>,
    );
    expect(wrapper).toHaveDisplayName("article");
  });

  describe("onClick", () => {
    it("allows user to focus the card", () => {
      const wrapper = shallow(
        <Card onClick={() => {}}>
          <CardBlock>My Card</CardBlock>
        </Card>,
      );

      expect(wrapper).toHaveClassName("clickable");
      expect(wrapper.prop("tabIndex")).toBe("0");
    });

    it("calls the onClick method when clicking", () => {
      const mock = jest.fn();
      const wrapper = shallow(
        <Card onClick={mock}>
          <CardBlock>My Card</CardBlock>
        </Card>,
      );

      wrapper.simulate("click");
      expect(mock).toHaveBeenCalled();
    });

    it("calls the onClick method when user press enter", () => {
      const mock = jest.fn();
      const wrapper = shallow(
        <Card onClick={mock}>
          <CardBlock>My Card</CardBlock>
        </Card>,
      );

      wrapper.simulate("keypress", { key: "Enter" });
      expect(mock).toHaveBeenCalled();
    });

    it("calls the onClick method when user press space", () => {
      const mock = jest.fn();
      const wrapper = shallow(
        <Card onClick={mock}>
          <CardBlock>My Card</CardBlock>
        </Card>,
      );

      wrapper.simulate("keydown", { key: " " });
      expect(mock).not.toHaveBeenCalled();
      wrapper.simulate("keyup", { key: " " });
      expect(mock).toHaveBeenCalled();
    });
  });
});
