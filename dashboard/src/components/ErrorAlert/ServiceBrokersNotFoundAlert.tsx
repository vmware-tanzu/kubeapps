import * as React from "react";

import NotFoundErrorAlert from "./NotFoundErrorAlert";

class ServiceBrokersNotFoundAlert extends React.Component {
  public render() {
    return (
      <NotFoundErrorAlert header="No Service Brokers installed.">
        <p>
          Ask an administrator to install a compatible{" "}
          <a href="https://github.com/osbkit/brokerlist" target="_blank">
            Service Broker
          </a>{" "}
          to browse, provision and manage external services within Kubeapps.
        </p>
      </NotFoundErrorAlert>
    );
  }
}

export default ServiceBrokersNotFoundAlert;
