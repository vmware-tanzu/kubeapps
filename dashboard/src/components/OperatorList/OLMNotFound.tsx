import * as React from "react";

import { NotFoundErrorAlert } from "../ErrorAlert";

const OLMNotFound: React.SFC = () => {
  return (
    <NotFoundErrorAlert header="Operator Lifecycle Manager (OLM) not installed.">
      <div>
        <p>
          Ask an administrator to install the{" "}
          <a
            href="https://github.com/operator-framework/operator-lifecycle-manager"
            target="_blank"
          >
            Operator Lifecycle Manager
          </a>{" "}
          to browse, provision and manage Operators within Kubeapps.
        </p>
        <button className="button button-primary button-small">Install OLM</button>
      </div>
    </NotFoundErrorAlert>
  );
};

export default OLMNotFound;
