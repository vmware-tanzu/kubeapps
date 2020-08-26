import * as React from "react";
import { Link } from "react-router-dom";

import Alert from "components/js/Alert";
import * as url from "shared/url";

interface IOperatorNotSupportedProps {
  namespace: string;
}

function OperatorNotSupported(props: IOperatorNotSupportedProps) {
  return (
    <Alert theme="warning">
      <h5>Operators are supported on the default cluster only</h5>
      <p>
        Kubeapps' Operator support enables the addition of{" "}
        <Link to={url.app.operators.list("default", props.namespace)}>
          operators on the default cluster only
        </Link>
        .
      </p>
    </Alert>
  );
}

export default OperatorNotSupported;
