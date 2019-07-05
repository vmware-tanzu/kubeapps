import * as React from "react";
import { Lock } from "react-feather";

import { UnexpectedErrorAlert } from ".";
import { IRBACRole } from "../../shared/types";
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
      <UnexpectedErrorAlert
        title={
          <span>
            You don't have sufficient permissions to {action} in {namespaceText(namespace)}
          </span>
        }
        icon={Lock}
        showGenericMessage={false}
      >
        <div>
          <p>Ask your administrator for the following RBAC roles:</p>
          <ul className="error__permisions-list">
            {roles.map((r, i) => (
              <PermissionsListItem key={i} namespace={namespace} role={r} />
            ))}
          </ul>
          <p>
            See the documentation for more info on{" "}
            <a
              href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/access-control.md"
              target="_blank"
            >
              access control in Kubeapps
            </a>
            .
          </p>
        </div>
      </UnexpectedErrorAlert>
    );
  }
}

export default PermissionsErrorPage;
