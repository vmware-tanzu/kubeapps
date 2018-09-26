import { shallow } from "enzyme";
import * as React from "react";
import LoadingWrapper from "../../components/LoadingWrapper";

// Shared jest examples that checks that the provided component is rendering the Loading Wrapper
export default (args: any) => {
  const { component: Component, props, state } = args;
  const renderComponent = () => {
    const wrapper = shallow(<Component {...props} />);

    if (state) {
      wrapper.setState(state);
    }
    return wrapper;
  };

  describe("loading spinner", () => {
    it("renders a wrapper in loaded state = false", () => {
      const wrapper = renderComponent();

      const loadingWrapper = wrapper.find(LoadingWrapper);
      expect(wrapper.find(LoadingWrapper)).toExist();
      expect(loadingWrapper.props().loaded).toEqual(false);
    });

    it("matches the snapshot", () => {
      const wrapper = renderComponent();
      expect(wrapper).toMatchSnapshot();
    });
  });
};
