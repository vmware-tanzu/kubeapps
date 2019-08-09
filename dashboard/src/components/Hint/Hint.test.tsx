import { shallow } from "enzyme";
import * as React from "react";
import * as ReactTooltip from "react-tooltip";
import Hint from ".";

it("should render a Hint with additional props and children", () => {
  const wrapper = shallow(
    <Hint reactTooltipOpts={{ foo: "bar" }} id="foobar">
      <p>this is a test!</p>
    </Hint>,
  );
  const tooltip = wrapper.find(ReactTooltip);
  expect(tooltip).toExist();
  expect(tooltip.props()).toMatchObject({ foo: "bar" });
  expect(wrapper).toMatchSnapshot();
});
