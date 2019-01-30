import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";
import { ISecret } from "../../../shared/types";
import SecretItem from "./SecretItem";
import SecretItemDatum from "./SecretItemDatum";

describe("componentDidMount", () => {
  it("calls getSecret", () => {
    const mock = jest.fn();
    shallow(<SecretItem name="foo" getSecret={mock} />);
    expect(mock).toHaveBeenCalled();
  });
});

context("when fetching secrets", () => {
  [undefined, { isFetching: true }].forEach(secret => {
    itBehavesLike("aLoadingComponent", {
      component: SecretItem,
      props: {
        secret,
        getSecret: jest.fn(),
      },
    });
    it("displays the name of the Secret", () => {
      const wrapper = shallow(<SecretItem secret={secret} name="foo" getSecret={jest.fn()} />);
      expect(wrapper.text()).toContain("foo");
    });
  });
});

context("when there is an error fetching the Secret", () => {
  const secret = {
    error: new Error('secrets "foo" not found'),
    isFetching: false,
  };
  const wrapper = shallow(<SecretItem secret={secret} name="foo" getSecret={jest.fn()} />);

  it("diplays the Service name in the first column", () => {
    expect(
      wrapper
        .find("td")
        .first()
        .text(),
    ).toEqual("foo");
  });

  it("displays the error message in the second column", () => {
    expect(
      wrapper
        .find("td")
        .at(1)
        .text(),
    ).toContain('Error: secrets "foo" not found');
  });
});

context("when there is a valid Secret", () => {
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
  const kubeItem = {
    isFetching: false,
    item: secret,
  };

  it("renders the Secret name and type", () => {
    const wrapper = shallow(<SecretItem secret={kubeItem} name="foo" getSecret={jest.fn()} />);
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
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a SecretItemDatum component for each Secret key", () => {
    const wrapper = shallow(<SecretItem secret={kubeItem} name="foo" getSecret={jest.fn()} />);
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
  });
});

context("when there is an empty Secret", () => {
  const secret = {
    metadata: {
      name: "foo",
    },
    type: "Opaque",
  } as ISecret;
  const kubeItem = {
    isFetching: false,
    item: secret,
  };

  it("displays a message", () => {
    const wrapper = shallow(<SecretItem secret={kubeItem} name="foo" getSecret={jest.fn()} />);
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
