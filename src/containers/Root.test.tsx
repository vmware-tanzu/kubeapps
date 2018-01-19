import { shallow } from "enzyme";
import * as React from "react";
import Root from "./Root";

it("renders without crashing", () => {
  shallow(<Root />);
});
