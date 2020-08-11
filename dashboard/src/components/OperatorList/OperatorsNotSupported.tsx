import { MessageAlert } from "components/ErrorAlert";
import * as React from "react";
import { Link } from "react-router-dom";

import * as url from "shared/url";

interface IOperatorNotSupportedProps {
  namespace: string;
}

function OperatorNotSupported(props: IOperatorNotSupportedProps) {
  return (
    <MessageAlert header="Operators are supported on the default cluster only">
      <div>
        <p className="margin-v-normal">
          Kubeapps' Operator support enables the addition of{" "}
          <Link to={url.app.operators.list("default", props.namespace)}>
            operators on the default cluster only
          </Link>
          .
        </p>
      </div>
    </MessageAlert>
  );
}

export default OperatorNotSupported;
