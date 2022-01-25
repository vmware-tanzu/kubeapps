// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";

// Shared jest examples that checks that the provided component is rendering the Loading Wrapper
/* eslint-disable jest/no-export */
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

      const loadingWrapper = wrapper.find("LoadingWrapper");
      expect(wrapper.find("LoadingWrapper")).toExist();
      expect(loadingWrapper.prop("loaded")).toEqual(false);
    });

    it("matches the snapshot", () => {
      const wrapper = renderComponent();
      expect(wrapper).toMatchSnapshot();
    });
  });
};
