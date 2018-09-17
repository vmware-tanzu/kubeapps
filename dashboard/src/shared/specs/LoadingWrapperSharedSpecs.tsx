import { shallow } from "enzyme";
import * as React from "react";

// Shared jest examples that checks that the provided component is rendering the Loading Wrapper
export default (Component: any, props: any) => {
  describe("loading spinner", () => {
    it("renders", () => {
      const wrapper = shallow(<Component {...props} />);
      expect(wrapper.find("LoadingWrapper")).toExist();
    });

    it("matches the snapshot", () => {
      const wrapper = shallow(<Component {...props} />);
      expect(wrapper).toMatchSnapshot();
    });
  });
};
