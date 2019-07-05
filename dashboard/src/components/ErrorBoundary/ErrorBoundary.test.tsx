import { mount } from "enzyme";
import * as React from "react";
import ErrorBoundary from ".";
import { UnexpectedErrorAlert } from "../ErrorAlert";

// tslint:disable:no-console
const consoleOrig = console.error;

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
      <ErrorBoundary>
        <BadRenderor throwError={true} />
      </ErrorBoundary>,
    );

    // Shows a generic error message
    const errorMessage = wrapper.find(UnexpectedErrorAlert);
    expect(errorMessage).toExist();
    expect(errorMessage.props().showGenericMessage).toBe(true);

    // Sets the internal state
    expect(wrapper.state("error")).toEqual(exampleError);
    const errorInfo: React.ErrorInfo = wrapper.state("errorInfo");
    expect(errorInfo.componentStack.length).not.toEqual(0);

    // console.error is called
    expect(console.error).toHaveBeenCalled();
  });

  it("renders only the wrapped components if no error", () => {
    const wrapper = mount(
      <ErrorBoundary>
        <BadRenderor />
      </ErrorBoundary>,
    );

    // shows the children component
    expect(wrapper.find(".no-error")).toExist();

    // Does not show a error message
    const errorMessage = wrapper.find(UnexpectedErrorAlert);
    expect(errorMessage).not.toExist();

    // the state is null
    expect(wrapper.state()).toEqual({ error: null, errorInfo: null });

    // console.error is not called
    expect(console.error).not.toHaveBeenCalled();
  });
});
