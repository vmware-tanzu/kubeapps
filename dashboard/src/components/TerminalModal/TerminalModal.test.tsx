import { shallow } from "enzyme";
import * as React from "react";

import TerminalModal from "./TerminalModal";

it("renders the message details modal", () => {
  const wrapper = shallow(
    <TerminalModal
      modalIsOpen={false}
      title={"This is a title"}
      message={"This is a message!"}
      closeModal={jest.fn()}
    />,
  );

  expect(wrapper.find(".Terminal__Top__Title").text()).toBe("This is a title");
  expect(wrapper.find(".Terminal__Code").text()).toBe("This is a message!");
  expect(wrapper).toMatchSnapshot();
});
