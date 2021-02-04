import { shallow } from "enzyme";
import Differential from "./Differential";

it("should render a diff between two strings", () => {
  const wrapper = shallow(
    <Differential title="test" oldValues="foo" newValues="bar" emptyDiffText="empty" />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("should print the emptyDiffText if there are no changes", () => {
  const wrapper = shallow(
    <Differential title="test" oldValues="foo" newValues="foo" emptyDiffText="No differences!" />,
  );
  expect(wrapper.text()).toMatch("No differences!");
  expect(wrapper.text()).not.toMatch("foo");
});
