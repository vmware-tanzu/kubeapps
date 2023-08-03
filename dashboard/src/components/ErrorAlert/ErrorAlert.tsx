// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import AlertGroup from "components/AlertGroup";
import RpcErrorMessage from "components/RpcErrorMessage";
import { ReactNode } from "react";
import { RpcError } from "shared/RpcError";
import { CustomError } from "shared/types";
import "./ErrorAlert.css";

export interface IErrorAlert {
  error: CustomError | Error;
  children?: React.ReactChildren | React.ReactNode | string;
}

function createWrap(message: any, index: number, indented: boolean): JSX.Element {
  return (
    <div className={indented ? "error-alert-indent" : "error-alert"} key={index}>
      {message}
    </div>
  );
}

function buildMessages(errors: Error[]): JSX.Element[] {
  return errors.map((cause, index) => {
    if (cause instanceof RpcError) {
      return createWrap(<RpcErrorMessage>{cause}</RpcErrorMessage>, index + 1, true);
    } else {
      return createWrap(cause.message, index + 1, true);
    }
  });
}

// Extension of Alert component for showing more meaningful Errors
export default function ErrorAlert({ error, children }: IErrorAlert) {
  let messages: ReactNode[];
  if (error instanceof CustomError) {
    messages = [createWrap(error.message, 0, false)];
    if (error.causes) {
      messages.push(buildMessages(error.causes));
    }
  } else if (error instanceof Error) {
    messages = [createWrap(error.message, 0, false)];
  } else {
    messages = [error];
  }
  return (
    <AlertGroup status="danger">
      {messages}
      {children}
    </AlertGroup>
  );
}
