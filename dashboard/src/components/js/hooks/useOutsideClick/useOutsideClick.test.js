// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import PropTypes from "prop-types";
import React, { useRef } from "react";
import { act } from "react-dom/test-utils";
import useOutsideClick from ".";

const TestComponent = ({ callback, enabled }) => {
  const ref = useRef(null);
  useOutsideClick(callback, ref, enabled);

  return <div ref={ref}>Test</div>;
};

TestComponent.propTypes = {
  callback: PropTypes.func,
  enabled: PropTypes.bool,
};

TestComponent.defaultProps = {
  enabled: true,
};

describe(useOutsideClick, () => {
  afterAll(() => {
    if (document.addEventListener) {
      document.addEventListener.mockRestore();
    }
  });

  it("should attach the event to the global event listener", () => {
    // Mock addEventListener
    const listeners = {};
    document.addEventListener = jest.fn(
      (event, cb) => {
        listeners[event] = cb;
      },
      { capture: true },
    );

    mount(<TestComponent />);
    expect(Object.keys(listeners).length).toBe(2);
    expect(listeners["mousedown"]).toBeDefined();
  });

  it("should attach the event only when enabled is true", async () => {
    // Mock addEventListener
    const listeners = {};
    document.addEventListener = jest.fn(
      (event, cb) => {
        listeners[event] = cb;
      },
      { capture: true },
    );

    const wrapper = mount(<TestComponent enabled={false} />);
    expect(Object.keys(listeners).length).toBe(0);

    await act(async () => {
      wrapper.setProps({ enabled: true });
    });
    // Force update
    wrapper.update();
    expect(Object.keys(listeners).length).toBe(1);
    expect(listeners["mousedown"]).toBeDefined();
  });

  /* eslint-disable jest/no-commented-out-tests */
  // TODO: Find a way to test real events
  // it('should run the callback when users click outside the element', () => {
  //   // This test is not implemented for now because we need to investigate
  //   // how to test global events. Most of the tooling about testing in React is
  //   // focused in Synthetic events, which are not enough to test this feature.
  // });
});
