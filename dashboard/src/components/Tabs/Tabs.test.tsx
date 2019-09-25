import { shallow } from "enzyme";
import * as React from "react";
import Tab from "./Tab";
import Tabs from "./Tabs";

const tabs = [{ title: "Foo", content: <div>foo</div> }, { title: "Bar", content: <div>bar</div> }];

it("renders a set of tabs", () => {
  const wrapper = shallow(<Tabs tabs={tabs} />);
  expect(wrapper).toMatchSnapshot();
});

it("shows by default the first tab", () => {
  const wrapper = shallow(<Tabs tabs={tabs} />);
  const tab = wrapper.find(Tab).filterWhere(t => t.prop("title") === tabs[0].title);
  expect(tab.prop("active")).toBe(true);
  const tabContent = wrapper.find("div").filterWhere(d => d.prop("id") === "tab-0-content");
  expect(tabContent.prop("hidden")).toBe(false);
});

it("shows the second tab content when clicked", () => {
  const wrapper = shallow(<Tabs tabs={tabs} />);
  let secondTab = wrapper.find(Tab).filterWhere(t => t.prop("title") === tabs[1].title);
  expect(secondTab.prop("active")).toBe(false);
  let tabContent = wrapper.find("div").filterWhere(d => d.prop("id") === "tab-1-content");
  expect(tabContent.prop("hidden")).toBe(true);

  secondTab.simulate("click");
  wrapper.update();

  secondTab = wrapper.find(Tab).filterWhere(t => t.prop("title") === tabs[1].title);
  tabContent = wrapper.find("div").filterWhere(d => d.prop("id") === "tab-1-content");
  expect(secondTab.prop("active")).toBe(true);
  expect(tabContent.prop("hidden")).toBe(false);
});
