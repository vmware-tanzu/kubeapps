// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

export class RpcError extends Error {
  // Following https://github.com/grpc/grpc-go/blob/master/internal/status/status.go#L140
  static RPC_ERROR_REGEX = /rpc error: code = (\w+) desc = (.*)/;
  static RPC_ERROR_CHECK = "rpc error:";

  readonly code: string;
  readonly desc: string;
  readonly message: string;

  constructor(error: Error) {
    super(error.message);
    Object.setPrototypeOf(this, new.target.prototype);
    [this.code, this.desc, this.message] = this.extractData(error.message);
  }
  public static isRpcError(error: Error) {
    return error.message.trim().indexOf(this.RPC_ERROR_CHECK) > -1;
  }
  private extractData(message: string): [string, string, string] {
    const found = message.trim().match(RpcError.RPC_ERROR_REGEX);
    return found
      ? [found[1], found[2], message.substring(0, message.indexOf(found[0]))]
      : ["", "", message];
  }
}
