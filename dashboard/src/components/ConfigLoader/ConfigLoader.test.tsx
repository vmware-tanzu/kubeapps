import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper";
import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";

import ConfigLoader from ".";

it("renders a loading wrapper", () => {
  const wrapper = shallow(<ConfigLoader loaded={false} getConfig={jest.fn()} />);
  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

context("when there is an error", () => {
  it("renders the error details", () => {
    const wrapper = shallow(
      <ConfigLoader error={new Error("Wrong config!")} getConfig={jest.fn()} />,
    );
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});

describe("componentDidMount", () => {
  it("calls getConfig", () => {
    const getConfig = jest.fn();
    mount(<ConfigLoader getConfig={getConfig} />);
    expect(getConfig).toHaveBeenCalled();
  });
});
