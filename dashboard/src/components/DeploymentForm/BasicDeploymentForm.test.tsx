import * as React from "react";

import { shallow } from "enzyme";
import { IBasicFormParam } from "shared/types";
import BasicDeploymentForm from "./BasicDeploymentForm";

const defaultProps = {
  params: [],
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
};

describe("username", () => {
  const param = {
    path: "wordpressUsername",
    value: "user",
  } as IBasicFormParam;

  it("renders a basic deployment with a username", () => {
    const onChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => onChange);
    const wrapper = shallow(
      <BasicDeploymentForm
        {...defaultProps}
        params={{ username: param }}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />,
    );
    wrapper.find("input#username").simulate("change");
    expect(handleBasicFormParamChange.mock.calls[0]).toEqual(["username", param]);
    expect(onChange.mock.calls.length).toBe(1);
  });
});
