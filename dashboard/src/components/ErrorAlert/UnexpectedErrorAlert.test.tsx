import { mount, shallow } from "enzyme";
import * as React from "react";

import ErrorPageHeader from "./ErrorAlertHeader";
import UnexpectedErrorAlert, { genericMessage } from "./UnexpectedErrorAlert";

describe("when text is passed", () => {
  const defaultProps = {
    text: "This is my error message",
  };
  it("renders the message in a paragraph unless raw", () => {
    const wrapper = shallow(<UnexpectedErrorAlert {...defaultProps} />);
    const messageWrapper = wrapper.find("p");
    expect(messageWrapper).toExist();
    expect(messageWrapper.text()).toContain(defaultProps.text);
  });

  it("renders the message in a terminal if raw", () => {
    const wrapper = shallow(<UnexpectedErrorAlert {...defaultProps} raw={true} />);
    const messageWrapper = wrapper.find(".Terminal");
    expect(messageWrapper).toExist();
    expect(messageWrapper.text()).toContain(defaultProps.text);
  });

  it("uses the passed title or fallbacks to the default one", () => {
    let wrapper = mount(<UnexpectedErrorAlert {...defaultProps} />);
    expect(wrapper.text()).toContain("Sorry! Something went wrong");

    wrapper = mount(<UnexpectedErrorAlert {...defaultProps} title="My error" />);
    expect(wrapper.text()).not.toContain("Sorry! Something went wrong");
    expect(wrapper.text()).toContain("My error");
  });

  it("renders the text even if the generic text is disabled", () => {
    const wrapper = shallow(<UnexpectedErrorAlert {...defaultProps} showGenericMessage={false} />);
    expect(
      wrapper
        .find(ErrorPageHeader)
        .shallow()
        .text(),
    ).toContain("Sorry! Something went wrong");
    const messageWrapper = wrapper.find("p");
    expect(messageWrapper.text()).toContain(defaultProps.text);
  });
});

describe("icon", () => {
  it("renders the default icon", () => {
    const wrapper = shallow(<UnexpectedErrorAlert />);
    const icon = wrapper.find(ErrorPageHeader).prop("icon") as any;
    expect(icon.name).toBe("X");
  });
  it("renders a custom icon", () => {
    const icon = <div>foo</div>;
    const wrapper = shallow(<UnexpectedErrorAlert icon={icon} />);
    const iconRendered = wrapper.find(ErrorPageHeader).prop("icon");
    expect(icon).toEqual(iconRendered);
  });
});

describe("genericMessage", () => {
  it("renders the default message", () => {
    const wrapper = shallow(<UnexpectedErrorAlert showGenericMessage={true} />);
    expect(wrapper.html()).toContain(shallow(genericMessage).html());
  });
  it("avoids the default message", () => {
    const wrapper = shallow(<UnexpectedErrorAlert showGenericMessage={false} />);
    expect(wrapper.html()).not.toContain(shallow(genericMessage).html());
  });
});

describe("children", () => {
  it("renders children components", () => {
    const wrapper = shallow(
      <UnexpectedErrorAlert>
        <div>more info!</div>
      </UnexpectedErrorAlert>,
    );
    expect(wrapper.find(".error__content").length).toBe(1);
    expect(wrapper.html()).toContain("more info!");
  });
  it("avoids extra elements", () => {
    const wrapper = shallow(<UnexpectedErrorAlert />);
    expect(wrapper.find(".error__content").length).toBe(0);
  });
});
