import { shallow } from "enzyme";
import * as React from "react";

import MessageDetails from "./MessageDetails";

it("renders the message details modal", () => {
  const wrapper = shallow(
    <MessageDetails modalIsOpen={false} message={"This is a message!"} closeModal={jest.fn()} />,
  );

  expect(wrapper.find(".Terminal__Code").text()).toBe("This is a message!");
  expect(wrapper).toMatchSnapshot();
});
