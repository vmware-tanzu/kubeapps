import { mount, shallow } from "enzyme";
import * as React from "react";

import UnexpectedErrorAlert from "./UnexpectedErrorAlert";

describe("when no text is passed", () => {
  it("renders a default troubleshooting info", () => {
    const wrapper = shallow(<UnexpectedErrorAlert />);
    expect(wrapper.text()).toContain("Troubleshooting");
    expect(wrapper.text()).toContain("Open an issue on GitHub");
    expect(wrapper).toMatchSnapshot();
  });
});

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
});
