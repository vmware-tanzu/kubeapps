import { shallow } from "enzyme";
import AppNotes from "./AppNotes";

it("renders AppNotes", () => {
  // Basic content testing
  expect(shallow(<AppNotes notes="" />).text()).toBe("");
  // eslint-disable-next-line redos/no-vulnerable
  expect(shallow(<AppNotes notes="Foo" />).text()).toMatch(/Notes.*Foo/);
});
