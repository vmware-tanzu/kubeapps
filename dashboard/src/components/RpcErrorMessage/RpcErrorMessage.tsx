// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { RpcError } from "shared/RpcError";
import "./RpcErrorMessage.css";

export interface IRpcErrorMessage {
  children: RpcError;
}

export default function RpcErrorMessage({ children }: IRpcErrorMessage) {
  return (
    <div className="error-alert-rpc">
      <span className="rpc-message">{children.message}</span>
      <ul>
        <li className="rpc-code">
          Code: <span className="rpc-value">{children.code}</span>
        </li>
        <li className="rpc-desc">
          Description: <span className="rpc-value">{children.desc}</span>
        </li>
      </ul>
    </div>
  );
}
