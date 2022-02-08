// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Alert from "components/js/Alert";
import { mount } from "enzyme";
import ErrorAlert from "./ErrorAlert";
import { CustomError } from "shared/types";

describe("Error Alert", () => {
  it("should render string errors", () => {
    const wrapper = mount(<ErrorAlert error={new Error("foo")} />);
    expect(wrapper.text()).toEqual("foo");
  });

  it("should handle empty string Error", () => {
    const error = new Error("");
    const wrapper = mount(<ErrorAlert error={error} />);
    expect(wrapper.text()).toEqual("");
  });

  it("should handle empty Error", () => {
    const error = new Error();
    const wrapper = mount(<ErrorAlert error={error} />);
    expect(wrapper.text()).toEqual("");
  });

  it("should render regular errors without breaklines", () => {
    const wrapper = mount(
      <ErrorAlert error={new Error("Error occurred: Another error message. Yet another msg.")} />,
    );
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.text()).toEqual("Error occurred: Another error message. Yet another msg.");
  });

  it("should render custom errors with plain messages", () => {
    const error = new CustomError(
      "An error occurred for tests: cause of the error. Another cause.",
    );
    const wrapper = mount(<ErrorAlert error={error} />);
    const alertTag = wrapper.find(Alert);
    expect(alertTag).toExist();
    expect(alertTag.text()).toEqual(
      "An error occurred for tests: cause of the error. Another cause.",
    );
  });

  it("should render custom errors with error causes", () => {
    const error = new CustomError("An error occurred for tests", [
      new Error("The first cause"),
      new Error("Second cause"),
    ]);
    const wrapper = mount(<ErrorAlert error={error} />);
    const alertTag = wrapper.find(Alert);
    expect(alertTag).toExist();
    expect(alertTag.find("div.error-alert")).toHaveLength(1);
    expect(alertTag.find("div.error-alert").text()).toBe("An error occurred for tests");
    expect(alertTag.find("div.error-alert-indent")).toHaveLength(2);
    expect(alertTag.find("div.error-alert-indent").at(0).text()).toBe("The first cause");
    expect(alertTag.find("div.error-alert-indent").at(1).text()).toBe("Second cause");
  });

  it("should render custom errors with error causes including rpc messages", () => {
    const error = new CustomError("An error occurred for tests", [
      new Error("The first cause"),
      new Error(
        "An error occurred when performing request: rpc error: code = Testing desc = The description of the RPC error",
      ),
      new Error("Even a third cause"),
    ]);
    const wrapper = mount(<ErrorAlert error={error} />);
    const alertTag = wrapper.find(Alert);
    expect(alertTag).toExist();
    expect(alertTag.find("div.error-alert")).toHaveLength(1);
    expect(alertTag.find("div.error-alert").text()).toBe("An error occurred for tests");
    expect(alertTag.find("div.error-alert-indent")).toHaveLength(3);
    expect(alertTag.find("div.error-alert-indent").at(0).text()).toBe("The first cause");
    expect(alertTag.find("div.error-alert-indent .error-alert-rpc")).toExist();
    expect(alertTag.find("div.error-alert-indent .error-alert-rpc .rpc-message").text()).toBe(
      "An error occurred when performing request: ",
    );
    expect(alertTag.find("div.error-alert-indent").at(2).text()).toBe("Even a third cause");
  });

  it("should render errors and children when present", () => {
    const wrapper = mount(
      <ErrorAlert error={new Error("Error message")}>
        <h1>Bang!</h1>
      </ErrorAlert>,
    );
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.text()).toEqual("Error messageBang!");
    expect(wrapper.find("h1")).toHaveLength(1);
    expect(wrapper.find("h1").text()).toBe("Bang!");
  });
});
