import { shallow } from "enzyme";
import * as React from "react";

import { Link } from "react-router-dom";
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

it("should generate a dummy link if it's not provided", () => {
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
  expect(wrapper.find(Link).props()).toMatchObject({ to: "#" });
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
