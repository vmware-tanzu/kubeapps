// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import Alert from "components/js/Alert";
import { mount } from "enzyme";
import React from "react";
import ErrorBoundary from "./ErrorBoundary";

/* eslint-disable no-console */

const consoleOrig = console.error;

const defaultProps = {
  logout: jest.fn(),
  children: <></>,
};

describe("ErrorBoundary around a component", () => {
  const exampleError = new Error("Bang!");
  const BadRenderor = (props: { throwError?: boolean }) => {
    if (props.throwError) {
      throw exampleError;
    }
    return <div className="no-error" />;
  };

  beforeEach(() => {
    // To avoid polluting the logs
    console.error = jest.fn();
  });

  afterEach(() => {
    console.error = consoleOrig;
  });

  it("captures any synchronous error thrown during a descendant render", () => {
    const wrapper = mount(
      <ErrorBoundary {...defaultProps}>
        <BadRenderor throwError={true} />
      </ErrorBoundary>,
    );

    // Shows a generic error message
    const errorMessage = wrapper.find(Alert);
    expect(errorMessage).toExist();
    expect(errorMessage.prop("theme")).toBe("danger");

    // Sets the internal state
    expect(wrapper.state("error")).toEqual(exampleError);
    const errorInfo: React.ErrorInfo = wrapper.state("errorInfo");
    expect(errorInfo.componentStack.length).not.toEqual(0);

    // console.error is called
    expect(console.error).toHaveBeenCalled();
  });

  it("renders only the wrapped components if no error", () => {
    const wrapper = mount(
      <ErrorBoundary {...defaultProps}>
        <BadRenderor />
      </ErrorBoundary>,
    );

    // shows the children component
    expect(wrapper.find(".no-error")).toExist();

    // Does not show a error message
    const errorMessage = wrapper.find(Alert);
    expect(errorMessage).not.toExist();

    // the state is null
    expect(wrapper.state()).toEqual({ error: null, errorInfo: null });

    // console.error is not called
    expect(console.error).not.toHaveBeenCalled();
  });
});

it("renders an error if it exists as a property", () => {
  const wrapper = mount(
    <ErrorBoundary {...defaultProps} error={new Error("boom!")}>
      <></>
    </ErrorBoundary>,
  );
  expect(wrapper.find(Alert).text()).toContain("boom!");
});

it("logs out when clicking on the link", () => {
  const logout = jest.fn();
  const wrapper = mount(
    <ErrorBoundary {...defaultProps} logout={logout} error={new Error("boom!")}>
      <></>
    </ErrorBoundary>,
  );
  const link = wrapper.find(Alert).find(CdsButton);
  expect(link).toExist();
  (link.prop("onClick") as any)();
  expect(logout).toHaveBeenCalled();
});

it("should not show the logout button if the error is catched from other component", () => {
  const wrapper = mount(<ErrorBoundary {...defaultProps} />);
  wrapper.setState({ error: new Error("Boom!") });
  const alert = wrapper.find(Alert);
  expect(alert).toExist();
  expect(alert.find(CdsButton)).not.toExist();
});
