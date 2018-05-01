import * as React from "react";
import { Lock } from "react-feather";

import { IRBACRole } from "../../shared/types";
import ErrorPageHeader from "./ErrorAlertHeader";
import { namespaceText } from "./helpers";
import PermissionsListItem from "./PermissionsListItem";

interface IPermissionsErrorPage {
  action: string;
  roles: IRBACRole[];
  namespace: string;
}

class PermissionsErrorPage extends React.Component<IPermissionsErrorPage> {
  public render() {
    const { action, namespace, roles } = this.props;
    return (
      <div className="alert alert-error margin-t-bigger" role="alert">
        <ErrorPageHeader icon={Lock}>
          You don't have sufficient permissions to {action} in {namespaceText(namespace)}.
        </ErrorPageHeader>
        <div className="error__content margin-l-enormous">
          <p>Ask your administrator for the following RBAC roles:</p>
          <ul className="error__permisions-list">
            {roles.map((r, i) => <PermissionsListItem key={i} namespace={namespace} role={r} />)}
          </ul>
          <p>
            See the documentation for more info on{" "}
            <a
              href="https://github.com/kubeapps/kubeapps/blob/master/docs/access-control.md"
              target="_blank"
            >
              access control in Kubeapps
            </a>.
          </p>
        </div>
      </div>
    );
  }
}

export default PermissionsErrorPage;
