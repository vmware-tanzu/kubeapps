import { shallow } from "enzyme";
import AppNotes from "./AppNotes";

it("renders AppNotes", () => {
  // Basic content testing
  expect(shallow(<AppNotes notes="" />).text()).toBe("");
  expect(shallow(<AppNotes notes="Foo" />).text()).toMatch(/Notes.*Foo/);
});
