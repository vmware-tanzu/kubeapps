import * as React from "react";
import { X } from "react-feather";

import ErrorPageHeader from "./ErrorAlertHeader";
import "./UnexpectedErrorAlert.css";

interface IUnexpectedErrorPage {
  raw?: boolean;
  text?: string;
  title?: string;
}

const genericMessage = (
  <div>
    <p>Troubleshooting:</p>
    <ul className="error__troubleshooting">
      <li>Check for network issues.</li>
      <li>Check your browser's JavaScript console for errors.</li>
      <li>
        Check the health of Kubeapps components{" "}
        <code>helm status &lt;kubeapps_release_name&gt;</code>.
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
  public static defaultProps: Partial<IUnexpectedErrorPage> = {
    title: "Sorry! Something went wrong.",
  };
  public render() {
    let message = genericMessage;
    if (this.props.text) {
      if (this.props.raw) {
        message = (
          <div className="error__content margin-l-enormous">
            <section className="Terminal terminal__error elevation-1 type-color-white">
              {this.props.text}
            </section>
          </div>
        );
      } else {
        message = <p>{this.props.text}</p>;
      }
    }
    // NOTE(miguel) We are using the non-undefined "!" token in `props.title` because our current version of
    // typescript does not support react's defaultProps and we are running it in strictNullChecks mode.
    // Newer versions of it seems to support it https://github.com/Microsoft/TypeScript/wiki/Roadmap#30-july-2018
    return (
      <div className="alert alert-error margin-t-bigger">
        <ErrorPageHeader icon={X}>{this.props.title!}</ErrorPageHeader>
        <div className="error__content margin-l-enormous">{message}</div>
      </div>
    );
  }
}

export default UnexpectedErrorPage;
