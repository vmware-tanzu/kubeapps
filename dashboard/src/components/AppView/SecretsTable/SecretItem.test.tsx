import { shallow } from "enzyme";
import * as React from "react";

import { ISecret } from "shared/types";
import SecretItem from "./SecretItem";

const secret = {
  apiVersion: "v1",
  kind: "Secret",
  type: "Opaque",
  metadata: {
    namespace: "ns",
    name: "deployment-one",
    annotations: "",
    creationTimestamp: "",
    selfLink: "",
    resourceVersion: "",
    uid: "",
  },
  data: { foo: "YmFy" }, // foo: bar
} as ISecret;

it("renders a secret (hidden by default)", () => {
  const wrapper = shallow(<SecretItem secret={secret} />);
  expect(wrapper.state()).toMatchObject({ showSecret: { foo: false } });
  expect(wrapper).toMatchSnapshot();
});

it("displays a secret when clicking on the icon", () => {
  const wrapper = shallow(<SecretItem secret={secret} />);
  expect(wrapper.state()).toMatchObject({ showSecret: { foo: false } });
  expect(wrapper.text()).toContain("foo:3 bytes");
  const icon = wrapper.find("a");
  expect(icon).toExist();
  icon.simulate("click");
  expect(wrapper.state()).toMatchObject({ showSecret: { foo: true } });
  expect(wrapper.text()).toContain("foo:bar");
});
