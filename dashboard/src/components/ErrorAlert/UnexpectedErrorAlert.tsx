import * as React from "react";

import ErrorPageHeader from "./ErrorAlertHeader";

class UnexpectedErrorPage extends React.Component {
  public render() {
    return (
      <div className="alert alert-error margin-t-bigger">
        <ErrorPageHeader>Sorry! Something went wrong.</ErrorPageHeader>
        <div className="error__content margin-l-enormous">
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
      </div>
    );
  }
}

export default UnexpectedErrorPage;
