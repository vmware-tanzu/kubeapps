import { CdsProgressCircle } from "@cds/react/progress-circle";
import { shallow } from "enzyme";
import LoadingWrapper from "./LoadingWrapper";

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

  it("does not render any children", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(ChildrenComponent)).not.toExist();
  });

  it("renders a progress circle", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(CdsProgressCircle)).toExist();
  });

  it("renders a mid size progress circle", () => {
    const wrapper = renderComponent({ ...props, size: "md" });
    expect(wrapper.find(CdsProgressCircle)).toExist();
    expect(wrapper.find(CdsProgressCircle).prop("size")).toBe("md");
  });

  it("renders a small progress circle", () => {
    const wrapper = renderComponent({ ...props, size: "sm" });
    expect(wrapper.find(CdsProgressCircle)).toExist();
    expect(wrapper.find(CdsProgressCircle).prop("size")).toBe("sm");
  });
});

describe("when loaded is true", () => {
  beforeEach(() => {
    props = {
      loaded: true,
    };
  });

  it("renders it wrapped component", () => {
    const wrapper = renderComponent(props);
    expect(wrapper.find(ChildrenComponent)).toExist();
  });
});
