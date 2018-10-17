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
