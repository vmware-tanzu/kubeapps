// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import olmLogo from "icons/olm-logo.svg";

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
          </a>
          .
        </p>
      </div>
    </div>
  );
}
