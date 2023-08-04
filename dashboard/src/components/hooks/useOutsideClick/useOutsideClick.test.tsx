// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import { mount } from "enzyme";
import { useRef } from "react";
import useOutsideClick from ".";

type TestComponentProps = {
  enabled: boolean;
};

const TestComponent = ({ enabled }: TestComponentProps) => {
  const ref = useRef(null);
  useOutsideClick(() => {}, [ref], enabled);

  return <div ref={ref}>Test</div>;
};

describe(useOutsideClick, () => {
  afterAll(() => {
    jest.restoreAllMocks();
  });

  it("should attach the event to the global event listener", () => {
    const listeners: { [key: string]: any } = {};

    // Mock addEventListener
    document.addEventListener = jest.fn((event, cb) => {
      listeners[event] = cb;
    });

    mount(<TestComponent enabled={true} />);

    expect(Object.keys(listeners).length).toBe(3);
    expect(listeners["mousedown"]).toBeDefined();
    expect(listeners["touchstart"]).toBeDefined();
  });

  it("should attach the event only when enabled is true", async () => {
    const listeners: { [key: string]: any } = {};

    // Mock addEventListener
    document.addEventListener = jest.fn((event, cb) => {
      listeners[event] = cb;
    });

    const wrapper = mount(<TestComponent enabled={false} />);

    expect(Object.keys(listeners).length).toBe(0);

    await act(async () => {
      wrapper.setProps({ enabled: true });
    });

    // Force update
    wrapper.update();
    expect(Object.keys(listeners).length).toBe(2);
    expect(listeners["mousedown"]).toBeDefined();
    expect(listeners["touchstart"]).toBeDefined();
  });
});
