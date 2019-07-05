import { shallow } from "enzyme";
import * as React from "react";

import Footer from "./Footer";

it("renders version and link to release", () => {
  const version = "v1.5.5-T500";
  const linkURI = `https://github.com/kubeapps/kubeapps/releases/tag/${version}`;

  const wrapper = shallow(<Footer appVersion={version} />);

  const link = wrapper.find("p.version-link a").first();
  expect(link.text()).toBe(version);
  expect(link.props().href).toBe(linkURI);
  expect(link).toMatchSnapshot();
});
