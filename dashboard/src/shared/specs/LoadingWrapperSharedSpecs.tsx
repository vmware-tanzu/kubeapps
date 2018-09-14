import { mount } from "enzyme";
import * as React from "react";
import LoadingSpinner from "../../components/LoadingSpinner";

// Shared jest examples that checks that the provided component is rendering the Loading Wrapper
export default (Component: any, props: any, state?: any) => {
  const renderComponent = () => {
    const wrapper = mount(<Component {...props} />);

    if (state) {
      wrapper.setState(state);
    }
    return wrapper;
  };

  describe("loading spinner", () => {
    it("renders", () => {
      const wrapper = renderComponent();

      expect(wrapper.find(LoadingSpinner)).toExist();
    });

    it("matches the snapshot", () => {
      const wrapper = renderComponent();
      expect(wrapper).toMatchSnapshot();
    });
  });
};
