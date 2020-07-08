import { shallow } from "enzyme";
import * as React from "react";
import PageHeader from "./PageHeader.v2";

it("should render a PageHeader", () => {
  const wrapper = shallow(
    <PageHeader>
      <h1>Title!</h1>
    </PageHeader>,
  );
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.text()).toContain("Title!");
});
