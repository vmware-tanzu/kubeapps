import Alert from "components/js/Alert";
import RpcErrorMessage from "components/RpcErrorMessage";
import { RpcError } from "shared/RpcError";
import { CustomError } from "shared/types";
import "./ErrorAlert.css";

export interface IErrorAlert {
  children: CustomError | Error | string;
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
export default function ErrorAlert({ children }: IErrorAlert) {
  let messages: any[];
  if (children instanceof CustomError) {
    messages = [createWrap(children.message, 0, false)];
    if (children.causes) {
      messages.push(buildMessages(children.causes));
    }
  } else if (children instanceof Error) {
    messages = [createWrap(children.message, 0, false)];
  } else {
    messages = [children];
  }
  return <Alert theme="danger">{messages}</Alert>;
}
