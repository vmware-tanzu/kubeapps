import * as React from "react";
import { Link } from "react-router-dom";

import NotFoundErrorAlert from "./NotFoundErrorAlert";

class ServiceCatalogNotInstalledAlert extends React.Component {
  public render() {
    return (
      <NotFoundErrorAlert header="Service Catalog not installed.">
        <div>
          <p>
            Ask an administrator to install the{" "}
            <a href="https://github.com/kubernetes-incubator/service-catalog" target="_blank">
              Kubernetes Service Catalog
            </a>{" "}
            to browse, provision and manage external services within Kubeapps.
          </p>
          <Link className="button button-primary button-small" to={"/charts/svc-cat/catalog"}>
            Install Catalog
          </Link>
        </div>
      </NotFoundErrorAlert>
    );
  }
}

export default ServiceCatalogNotInstalledAlert;
