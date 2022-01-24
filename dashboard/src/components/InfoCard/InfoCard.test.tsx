// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Icon from "components/Icon/Icon";
import { shallow } from "enzyme";
import { Link } from "react-router-dom";
import { CardBlock } from "../js/Card";
import InfoCard from "./InfoCard";

it("should render a Card", () => {
  const wrapper = shallow(
    <InfoCard
      title="foo"
      info="foobar"
      link="/a/link/somewhere"
      icon="an-icon.png"
      tag1Class="blue"
      tag1Content="database"
      tag2Class="red"
      tag2Content="running"
    />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("should generate a stub link if it's not provided", () => {
  const wrapper = shallow(
    <InfoCard
      title="foo"
      info="foobar"
      icon="an-icon.png"
      tag1Class="blue"
      tag1Content="database"
      tag2Class="red"
      tag2Content="running"
    />,
  );

  expect(wrapper.find(Link).prop("to")).toBe("#");
});

it("should avoid tags if they are not defined", () => {
  const wrapper = shallow(
    <InfoCard title="foo" info="foobar" link="/a/link/somewhere" icon="an-icon.png" />,
  );
  expect(wrapper.find(".ListItem__content__info_tag")).not.toExist();
});

it("should parse JSX elements in the tags", () => {
  const tag1 = <span className="tag1">tag1</span>;
  const tag2 = <span className="tag2">tag2</span>;
  const wrapper = shallow(
    <InfoCard
      title="foo"
      info="foobar"
      link="/a/link/somewhere"
      icon="an-icon.png"
      tag1Class="blue"
      tag1Content={tag1}
      tag2Class="red"
      tag2Content={tag2}
    />,
  );
  const tag1Elem = wrapper.find(".tag1");
  expect(tag1Elem).toExist();
  expect(tag1Elem.text()).toBe("tag1");
  const tag2Elem = wrapper.find(".tag2");
  expect(tag2Elem).toExist();
  expect(tag2Elem.text()).toBe("tag2");
});

it("should parse a description as text", () => {
  const wrapper = shallow(<InfoCard title="foo" info="foobar" description="a description" />);
  expect(wrapper.find(CardBlock).html()).toContain("a description");
});

it("should parse a description as JSX.Element", () => {
  const desc = <div className="description">This is a description</div>;
  const wrapper = shallow(<InfoCard title="foo" info="foobar" description={desc} />);
  expect(wrapper.find(".description").text()).toBe("This is a description");
});

it("should parse an icon", () => {
  const wrapper = shallow(<InfoCard title="foo" info="foobar" />);
  expect(wrapper.find(Icon)).toExist();
});

it("should parse a background img", () => {
  const wrapper = shallow(<InfoCard title="foo" info="foobar" bgIcon="img.png" />);
  expect(wrapper.find(".bg-img").find("img")).toExist();
});

it("should parse a tooltip component", () => {
  const wrapper = shallow(<InfoCard title="foo" info="foobar" tooltip={<div id="foo" />} />);
  expect(wrapper.find(".info-card-header").find("#foo")).toExist();
});
