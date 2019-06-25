import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { ISecret } from "../../../../../shared/types";
import SecretItem from "./SecretItem";
import SecretItemDatum from "./SecretItemDatum";

it("renders a SecretItemDatum component for each Secret key", () => {
  const secret = {
    metadata: {
      name: "foo",
    },
    type: "Opaque",
    data: {
      foo: "YmFy", // bar
      foo2: "YmFyMg==", // bar2
    } as { [s: string]: string },
  } as ISecret;
  const wrapper = shallow(<SecretItem resource={secret} />);
  expect(
    wrapper
      .find("td")
      .at(2)
      .find(SecretItemDatum),
  ).toHaveLength(2);
  expect(
    wrapper
      .find(SecretItemDatum)
      .first()
      .props(),
  ).toMatchObject({ name: "foo", value: "YmFy" });
  expect(wrapper).toMatchSnapshot();
});

context("when there is an empty Secret", () => {
  const secret = {
    metadata: {
      name: "foo",
    },
    type: "Opaque",
  } as ISecret;
  it("displays a message", () => {
    const wrapper = shallow(<SecretItem resource={secret} />);
    expect(
      wrapper
        .find("td")
        .at(0)
        .text(),
    ).toContain("foo");
    expect(
      wrapper
        .find("td")
        .at(1)
        .text(),
    ).toContain("Opaque");
    expect(
      wrapper
        .find("td")
        .at(2)
        .text(),
    ).toContain("This Secret is empty");
    expect(wrapper).toMatchSnapshot();
  });
});
