// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

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
    const view = renderComponent(props);
    expect(view.find(ChildrenComponent)).not.toExist();
  });

  it("renders a progress circle", () => {
    const view = renderComponent(props);
    expect(view.find(CdsProgressCircle)).toExist();
  });

  it("renders a mid size progress circle", () => {
    const view = renderComponent({ ...props, size: "md" });
    expect(view.find(CdsProgressCircle)).toExist();
    expect(view.find(CdsProgressCircle).prop("size")).toBe("md");
  });

  it("renders a small progress circle", () => {
    const view = renderComponent({ ...props, size: "sm" });
    expect(view.find(CdsProgressCircle)).toExist();
    expect(view.find(CdsProgressCircle).prop("size")).toBe("sm");
  });
});

describe("when loaded is true", () => {
  beforeEach(() => {
    props = {
      loaded: true,
    };
  });

  it("renders it wrapped component", () => {
    const view = renderComponent(props);
    expect(view.find(ChildrenComponent)).toExist();
  });
});
