import * as React from "react";
import { Link } from "react-router-dom";

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
        <Link className="button button-primary button-small" to={"/charts/svc-cat/catalog"}>
          Install OLM
        </Link>
      </div>
    </NotFoundErrorAlert>
  );
};

export default OLMNotFound;
