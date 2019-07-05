import * as React from "react";
import { X } from "react-feather";

import ErrorPageHeader from "./ErrorAlertHeader";
import "./UnexpectedErrorAlert.css";

interface IUnexpectedErrorPage {
  icon?: any;
  raw?: boolean;
  showGenericMessage?: boolean;
  text?: string;
  title?: string | JSX.Element;
}

export const genericMessage = (
  <div>
    <p>Troubleshooting:</p>
    <ul className="error__troubleshooting">
      <li>Check for network issues.</li>
      <li>Check your browser's JavaScript console for additional errors.</li>
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
    showGenericMessage: false,
  };
  public render() {
    let customMessage = null;
    if (this.props.text) {
      if (this.props.raw) {
        customMessage = (
          <div className="error__content">
            <section className="Terminal terminal__error elevation-1 type-color-white">
              {this.props.text}
            </section>
          </div>
        );
      } else {
        customMessage = <p>{this.props.text}</p>;
      }
    }
    // NOTE(miguel) We are using the non-undefined "!" token in `props.title` because our current version of
    // typescript does not support react's defaultProps and we are running it in strictNullChecks mode.
    // Newer versions of it seems to support it https://github.com/Microsoft/TypeScript/wiki/Roadmap#30-july-2018
    return (
      <div className="alert alert-error margin-v-bigger">
        <ErrorPageHeader icon={this.props.icon || X}>{this.props.title!}</ErrorPageHeader>
        {customMessage && (
          <div className="error__content margin-l-enormous margin-b-big">{customMessage}</div>
        )}
        {this.props.showGenericMessage && (
          <div className="error__content margin-l-enormous">{genericMessage}</div>
        )}
        {this.props.children && (
          <div className="error__content margin-l-enormous">{this.props.children}</div>
        )}
      </div>
    );
  }
}

export default UnexpectedErrorPage;
