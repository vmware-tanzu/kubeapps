import { shallow } from "enzyme";
import * as React from "react";
import LoadingWrapper, { LoaderType } from ".";
import LoadingSpinner from "../LoadingSpinner";
import LoadingPlaceholder from "./LoadingPlaceholder";

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

  it("renders a spinner", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(LoadingSpinner)).toExist();
  });

  it("does not render any children", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(ChildrenComponent)).not.toExist();
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

  it("does not renders a spinner", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(LoadingSpinner)).not.toExist();
  });

  it("renders it wrapped component", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(ChildrenComponent)).toExist();
  });
});

describe("loader types", () => {
  it("renders the LoadingPlaceholder type when specified", () => {
    const wrapper = renderComponent({ type: LoaderType.Placeholder });
    expect(wrapper).toMatchSnapshot();
    expect(wrapper.find(LoadingPlaceholder)).toExist();
  });
});
