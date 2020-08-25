import olmLogo from "icons/olm-big.svg";
import * as React from "react";

export default function OLMNotFound() {
  return (
    <div className="section-not-found">
      <div>
        <img src={olmLogo} alt="olm-log" />
        <h4>The Operator Lifecycle Manager (OLM) is not available</h4>
        <p className="section-description">
          Ask an administrator to install the{" "}
          <a
            href="https://github.com/operator-framework/operator-lifecycle-manager"
            target="_blank"
            rel="noopener noreferrer"
          >
            Operator Lifecycle Manager
          </a>{" "}
          to browse, provision and manage Operators within Kubeapps. <br />
          To install the OLM, check{" "}
          <a
            href="https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md"
            target="_blank"
            rel="noopener noreferrer"
          >
            the installation instructions
          </a>{" "}
          or execute the following command in a terminal with <code>kubectl</code> available and
          configured:
        </p>
        <section className="terminal-wrapper">
          <pre className="terminal-code">
            curl -L
            https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.15.1/install.sh
            -o install.sh <br />
            chmod +x install.sh <br />
            ./install.sh 0.15.1
          </pre>
        </section>
      </div>
    </div>
  );
}
