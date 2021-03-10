import { shallow } from "enzyme";
import AppValues from "./AppValues";

it("match snapshot with values", () => {
  const wrapper = shallow(<AppValues values="foo: bar" />);
  expect(wrapper).toMatchSnapshot();
});

it("match snapshot without values", () => {
  const wrapper = shallow(<AppValues values="" />);
  expect(wrapper).toMatchSnapshot();
});
