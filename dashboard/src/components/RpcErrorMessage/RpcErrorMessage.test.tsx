// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import { RpcError } from "shared/RpcError";
import RpcErrorMessage from ".";

describe("RPC Error message", () => {
  it("should render rpc error", () => {
    const error = new RpcError(
      new Error(
        "An error occurred for tests: rpc error: code = Testing desc = The description of the RPC error",
      ),
    );
    const wrapper = mount(<RpcErrorMessage>{error}</RpcErrorMessage>);
    expect(wrapper.find("span.rpc-message").text()).toContain("An error occurred for tests: ");
    expect(wrapper.find("li.rpc-code .rpc-value").text()).toContain("Testing");
    expect(wrapper.find("li.rpc-desc .rpc-value").text()).toContain(
      "The description of the RPC error",
    );
  });

  it("should render message with empty details", () => {
    const error = new RpcError(new Error("Non RPC error message text"));
    const wrapper = mount(<RpcErrorMessage>{error}</RpcErrorMessage>);
    expect(wrapper.find("span.rpc-message").text()).toContain("Non RPC error message text");
    expect(wrapper.find("li.rpc-code .rpc-value").text()).toContain("");
    expect(wrapper.find("li.rpc-desc .rpc-value").text()).toContain("");
  });
});
