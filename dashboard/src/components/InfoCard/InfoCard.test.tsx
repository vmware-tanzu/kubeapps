// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Icon from "components/Icon/Icon";
import { Link } from "react-router-dom";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import InfoCard, { IInfoCardProps } from "./InfoCard";

const defaultProps: IInfoCardProps = {
  title: "foo",
  info: "foobar",
  icon: "an-icon.png",
  description: "a description",
  tag1Class: "blue",
  tag1Content: "database",
  tag2Class: "red",
  tag2Content: "running",
};

describe(InfoCard, () => {
  describe("Links", () => {
    it("should generate a stub link if it's not provided", () => {
      const wrapper = mountWrapper(defaultStore, <InfoCard {...defaultProps} />);

      expect(wrapper.find(Link).prop("to")).toBe("#");
    });

    it("should avoid tags if they are not defined", () => {
      const props: IInfoCardProps = {
        ...defaultProps,
        link: "/a/link/somewhere",
      };
      const wrapper = mountWrapper(defaultStore, <InfoCard {...props} />);

      expect(wrapper.find(".ListItem__content__info_tag")).not.toExist();
    });
  });

  describe("Card Header", () => {
    it("includes the expected CSS class and renders the content correctly", () => {
      const wrapper = mountWrapper(defaultStore, <InfoCard {...defaultProps} />);

      const headerTitle = wrapper.find(".card-title");
      expect(headerTitle).toExist();
      expect(headerTitle).toHaveText(defaultProps.title);
    });

    it("should parse a tooltip component", () => {
      const props: IInfoCardProps = {
        ...defaultProps,
        title: "foo",
        info: "foobar",
        tooltip: <div id="foo" />,
      };
      const wrapper = mountWrapper(defaultStore, <InfoCard {...props} />);

      expect(wrapper.find(".info-card-header").find("#foo")).toExist();
    });
  });

  describe("Card Block", () => {
    it("renders the content correctly and renders the content correctly", () => {
      const wrapper = mountWrapper(defaultStore, <InfoCard {...defaultProps} />);

      const block = wrapper.find(".info-card-block");
      expect(block).toExist();
      expect(block.html()).toContain(defaultProps.description);
    });

    it("should parse a description as text", () => {
      const props: IInfoCardProps = {
        ...defaultProps,
        title: "foo",
        info: "foobar",
        description: "a description",
      };
      const wrapper = mountWrapper(defaultStore, <InfoCard {...props} />);

      expect(wrapper.find(".card-description").html()).toContain("a description");
    });

    it("should parse a description as JSX.Element", () => {
      const props: IInfoCardProps = {
        ...defaultProps,
        title: "foo",
        info: "foobar",
        description: <div className="my-description">This is a description</div>,
      };
      const wrapper = mountWrapper(defaultStore, <InfoCard {...props} />);

      expect(wrapper.find(".my-description").text()).toBe("This is a description");
    });

    it("should parse an icon", () => {
      const props: IInfoCardProps = {
        ...defaultProps,
        title: "foo",
        info: "foobar",
      };
      const wrapper = mountWrapper(defaultStore, <InfoCard {...props} />);

      expect(wrapper.find(Icon)).toExist();
    });

    it("should parse a background img", () => {
      const props: IInfoCardProps = {
        ...defaultProps,
        title: "foo",
        info: "foobar",
        bgIcon: "img.png",
      };
      const wrapper = mountWrapper(defaultStore, <InfoCard {...props} />);

      expect(wrapper.find(".bg-img").find("img")).toExist();
    });
  });

  describe("Card Footer", () => {
    it("renders the content correctly and renders the content correctly", () => {
      const wrapper = mountWrapper(defaultStore, <InfoCard {...defaultProps} />);

      const footer = wrapper.find(".card-footer");
      expect(footer).toExist();
      expect(footer.html()).toContain(defaultProps.info.toString());
    });

    it("should parse JSX elements in the tags", () => {
      const props: IInfoCardProps = {
        ...defaultProps,
        link: "/a/link/somewhere",
        tag1Content: <span className="tag1">tag1</span>,
        tag2Content: <span className="tag2">tag2</span>,
      };
      const wrapper = mountWrapper(defaultStore, <InfoCard {...props} />);

      const tag1Elem = wrapper.find(".tag1");
      expect(tag1Elem).toExist();
      expect(tag1Elem.text()).toBe("tag1");

      const tag2Elem = wrapper.find(".tag2");
      expect(tag2Elem).toExist();
      expect(tag2Elem.text()).toBe("tag2");
    });
  });
});
