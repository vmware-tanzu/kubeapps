import * as React from "react";
import { X } from "react-feather";

import ErrorPageHeader from "./ErrorAlertHeader";

interface IUnexpectedErrorPage {
  raw?: boolean;
  text?: string;
}

const genericMessage = (
  <div>
    <p>Troubleshooting:</p>
    <ul className="error__troubleshooting">
      <li>Check for network issues.</li>
      <li>Check your browser's JavaScript console for errors.</li>
      <li>
        Check the health of Kubeapps components{" "}
        <code>kubectl get po --all-namespaces -l created-by=kubeapps</code>.
      </li>
      <li>
        <a href="https://github.com/kubeapps/kubeapps/issues/new" target="_blank">
          Open an issue on GitHub
        </a>{" "}
        if you think you've encountered a bug.
      </li>
    </ul>
  </div>
);

class UnexpectedErrorPage extends React.Component<IUnexpectedErrorPage> {
  public render() {
    let message = genericMessage;
    if (this.props.text) {
      if (this.props.raw) {
        message = (
          <div className="error__content margin-l-enormous">
            <section
              className="Terminal elevation-1 type-color-white"
              style={{ wordBreak: "break-all" }}
            >
              {this.props.text}
            </section>
          </div>
        );
      } else {
        message = <p>{this.props.text}</p>;
      }
    }
    return (
      <div className="alert alert-error margin-t-bigger">
        <ErrorPageHeader icon={X}>Sorry! Something went wrong.</ErrorPageHeader>
        <div className="error__content margin-l-enormous">{message}</div>
      </div>
    );
  }
}

export default UnexpectedErrorPage;
