import { shallow } from "enzyme";
import * as React from "react";

import Card from "./Card";

describe("cssClass", () => {
  it("should render with the default class", () => {
    const wrapper = shallow(<Card />);
    expect(wrapper.find(".Card").exists()).toBe(true);
    expect(wrapper).toMatchSnapshot();
  });

  it("should use the className property", () => {
    const wrapper = shallow(<Card className="foo" />);
    expect(wrapper.find(".foo").exists()).toBe(true);
  });

  it("should add the responsive class", () => {
    const wrapper = shallow(<Card responsive={true} />);
    expect(wrapper.find(".Card-responsive").exists()).toBe(true);
  });

  it("should add the responsive class with colums", () => {
    const wrapper = shallow(<Card responsive={true} responsiveColumns={1} />);
    expect(wrapper.find(".Card-responsive-1").exists()).toBe(true);
  });
});

it("should render the children elements", () => {
  const wrapper = shallow(
    <Card>
      <div>foo</div>
    </Card>,
  );
  expect(wrapper.text()).toContain("foo");
});
