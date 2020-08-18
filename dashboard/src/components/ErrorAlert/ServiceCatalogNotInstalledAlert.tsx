import * as React from "react";

import NotFoundErrorAlert from "./NotFoundErrorAlert";

class ServiceCatalogNotInstalledAlert extends React.Component {
  public render() {
    return (
      <NotFoundErrorAlert header="Service Catalog not installed.">
        <div>
          <p>
            Ask an administrator to install the{" "}
            <a
              href="https://github.com/kubernetes-incubator/service-catalog"
              target="_blank"
              rel="noopener noreferrer"
            >
              Kubernetes Service Catalog
            </a>{" "}
            to browse, provision and manage external services within Kubeapps.
          </p>
        </div>
      </NotFoundErrorAlert>
    );
  }
}

export default ServiceCatalogNotInstalledAlert;
