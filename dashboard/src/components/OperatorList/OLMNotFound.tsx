import * as React from "react";

import { NotFoundErrorAlert } from "../ErrorAlert";

class OLMNotFound extends React.Component {
  public render() {
    return (
      <NotFoundErrorAlert header="Operator Lifecycle Manager (OLM) not installed.">
        <div>
          <p>
            Ask an administrator to install the{" "}
            <a
              href="https://github.com/operator-framework/operator-lifecycle-manager"
              target="_blank"
              rel="noopener noreferrer"
            >
              Operator Lifecycle Manager
            </a>{" "}
            to browse, provision and manage Operators within Kubeapps.
          </p>
          To install the OLM, execute the following command in a terminal with <code>kubectl</code>{" "}
          available and configured:
          <section className="AppNotes Terminal elevation-1 margin-v-big">
            <div className="Terminal__Top type-small">
              <div className="Terminal__Top__Buttons">
                <span className="Terminal__Top__Button Terminal__Top__Button--red" />
                <span className="Terminal__Top__Button Terminal__Top__Button--yellow" />
                <span className="Terminal__Top__Button Terminal__Top__Button--green" />
              </div>
            </div>
            <div className="Terminal__Tab">
              <pre className="Terminal__Code">
                <code>
                  curl -L
                  https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.15.1/install.sh
                  -o install.sh <br />
                  chmod +x install.sh <br />
                  ./install.sh 0.15.1
                </code>
              </pre>
            </div>
          </section>
        </div>
      </NotFoundErrorAlert>
    );
  }
}

export default OLMNotFound;
