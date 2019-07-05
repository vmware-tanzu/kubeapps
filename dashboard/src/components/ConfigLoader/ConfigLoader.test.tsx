import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import ConfigLoader from ".";
import UnexpectedErrorPage from "../../components/ErrorAlert/UnexpectedErrorAlert";
import itBehavesLike from "../../shared/specs";

context("when the config is not ready", () => {
  itBehavesLike("aLoadingComponent", {
    component: ConfigLoader,
    props: {
      loaded: false,
      getConfig: jest.fn(),
    },
  });
});

context("when there is an error", () => {
  it("renders the error details", () => {
    const wrapper = shallow(
      <ConfigLoader error={new Error("Wrong config!")} getConfig={jest.fn()} />,
    );
    expect(wrapper.find(UnexpectedErrorPage)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});

describe("componentDidMount", () => {
  it("calls getConfig", () => {
    const getConfig = jest.fn();
    shallow(<ConfigLoader getConfig={getConfig} />);
    expect(getConfig).toHaveBeenCalled();
  });
});
