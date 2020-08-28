import { MessageAlert } from "components/ErrorAlert";
import * as React from "react";
import { Link } from "react-router-dom";

import * as url from "shared/url";

interface IOperatorNotSupportedProps {
  kubeappsCluster: string;
  namespace: string;
}

function OperatorNotSupported(props: IOperatorNotSupportedProps) {
  if (props.kubeappsCluster) {
    return (
      <MessageAlert header="Operators are supported on the cluster on which Kubeapps is installed only">
        <div>
          <p className="margin-v-normal">
            Kubeapps' Operator support enables the addition of{" "}
            <Link to={url.app.operators.list(props.kubeappsCluster, props.namespace)}>
              operators on the cluster on which Kubeapps is installed only
            </Link>
            .
          </p>
        </div>
      </MessageAlert>
    );
  } else {
    return (
      <MessageAlert header="Operators are not supported on this installation">
        <div>
          <p className="margin-v-normal">
            Kubeapps' Operator support enables the addition of operators on the cluster on which
            Kubeapps is installed only. This installation of Kubeapps is configured without access
            to that cluster.
          </p>
        </div>
      </MessageAlert>
    );
  }
}

export default OperatorNotSupported;
