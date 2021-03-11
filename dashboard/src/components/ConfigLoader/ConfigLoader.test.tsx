import Alert from "components/js/Alert";
import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";

import ConfigLoader from ".";
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
