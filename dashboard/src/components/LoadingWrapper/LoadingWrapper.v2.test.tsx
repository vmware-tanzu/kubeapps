import { shallow } from "enzyme";
import * as React from "react";
import Spinner from "../js/Spinner";
import LoadingWrapper from "./LoadingWrapper.v2";

let props = {} as any;

const ChildrenComponent = () => <div>Hello dad!</div>;

const renderComponent = (p: any) => {
  return shallow(
    <LoadingWrapper {...p}>
      <ChildrenComponent />
    </LoadingWrapper>,
  );
};

describe("when loaded is false", () => {
  beforeEach(() => {
    props = {
      loaded: false,
    };
  });

  it("matches the snapshot", () => {
    const wrapper = renderComponent(props);
    expect(wrapper).toMatchSnapshot();
  });

  it("does not render any children", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(ChildrenComponent)).not.toExist();
  });

  it("renders a Spinner", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(Spinner)).toExist();
  });

  it("renders a mid size Spinner", () => {
    const wrapper = renderComponent({ ...props, medium: true });
    expect(wrapper.find(Spinner)).toExist();
    expect(wrapper.find(Spinner).prop("medium")).toBe(true);
  });

  it("renders a small Spinner", () => {
    const wrapper = renderComponent({ ...props, small: true });
    expect(wrapper.find(Spinner)).toExist();
    expect(wrapper.find(Spinner).prop("small")).toBe(true);
  });
});

describe("when loaded is true", () => {
  beforeEach(() => {
    props = {
      loaded: true,
    };
  });

  it("matches the snapshot", () => {
    const wrapper = renderComponent(props);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders it wrapped component", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(ChildrenComponent)).toExist();
  });
});
