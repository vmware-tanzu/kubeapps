// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import { mount } from "enzyme";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import Card from "./Card";
import CardBlock from "./CardBlock";
import CardFooter from "./CardFooter";
import CardHeader from "./CardHeader";
import CardText from "./CardText";
import CardTitle from "./CardTitle";

describe(Card, () => {
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

  describe("onClick", () => {
    it("allows user to focus the card", () => {
      const wrapper = mountWrapper(
        defaultStore,
        <Card onClick={() => {}}>
          <CardBlock>My Card</CardBlock>
        </Card>,
      );

      expect(wrapper.find(Card).childAt(0)).toHaveClassName("clickable");
      expect(wrapper.find(Card).childAt(0).prop("tabIndex")).toBe(0);
    });

    it("calls the onClick method when clicking", () => {
      const mock = jest.fn();
      const wrapper = mountWrapper(
        defaultStore,
        <Card onClick={mock}>
          <CardBlock>My Card</CardBlock>
        </Card>,
      );

      act(() => {
        wrapper.find(Card).simulate("click");
      });

      expect(mock).toHaveBeenCalled();
    });

    it("calls the onClick method when user press enter", () => {
      const mock = jest.fn();
      const wrapper = mountWrapper(
        defaultStore,
        <Card onClick={mock}>
          <CardBlock>My Card</CardBlock>
        </Card>,
      );
      act(() => {
        wrapper.find(Card).simulate("keydown", { key: "Enter" });
      });

      expect(mock).toHaveBeenCalled();
    });

    it("calls the onClick method when user press space", () => {
      const mock = jest.fn();
      const wrapper = mountWrapper(
        defaultStore,
        <Card onClick={mock}>
          <CardBlock>My Card</CardBlock>
        </Card>,
      );

      act(() => {
        wrapper.find(Card).simulate("keydown", { key: " " });
      });

      expect(mock).toHaveBeenCalled();
    });
  });
});
