// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { RpcError } from "./RpcError";

describe("RpcError", () => {
  it("parses correctly an Error", () => {
    const rpcError = new RpcError(
      new Error(
        "An error occurred for tests: rpc error: code = Testing desc = The description of the RPC error",
      ),
    );
    expect(rpcError.message).toBe("An error occurred for tests: ");
    expect(rpcError.code).toBe("Testing");
    expect(rpcError.desc).toBe("The description of the RPC error");
  });

  it("parses correctly a normal error", () => {
    const rpcError = new RpcError(new Error("An error occurred for tests: reason."));
    expect(rpcError.message).toBe("An error occurred for tests: reason.");
    expect(rpcError.code).toBe("");
    expect(rpcError.desc).toBe("");
  });

  it("handles correctly an empty string as param", () => {
    const rpcError = new RpcError(new Error(""));
    expect(rpcError.message).toBe("");
    expect(rpcError.code).toBe("");
    expect(rpcError.desc).toBe("");
  });

  it("handles correctly undefined param", () => {
    const rpcError = new RpcError(new Error(undefined));
    expect(rpcError.message).toBe("");
    expect(rpcError.code).toBe("");
    expect(rpcError.desc).toBe("");
  });
});
