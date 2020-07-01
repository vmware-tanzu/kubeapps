import * as React from "react";
import { Lock } from "react-feather";

import { UnexpectedErrorAlert } from ".";
import { IRBACRole } from "../../shared/types";
import { namespaceText } from "./helpers";
import PermissionsListItem from "./PermissionsListItem";

interface IPermissionsErrorPage {
  action: string;
  roles: IRBACRole[];
  rawMessage: string;
  namespace: string;
}

class PermissionsErrorPage extends React.Component<IPermissionsErrorPage> {
  public render() {
    const { action, namespace, roles, rawMessage } = this.props;
    return (
      <UnexpectedErrorAlert
        title={
          <span>
            You don't have sufficient permissions to {action}{" "}
            {namespace && <>in {namespaceText(namespace)}</>}
          </span>
        }
        icon={Lock}
        showGenericMessage={false}
      >
        <div>
          {roles.length > 0 ? (
            <>
              <p>Ask your administrator for the following RBAC roles:</p>
              <ul className="error__permisions-list">
                {roles.map((r, i) => (
                  <PermissionsListItem key={i} namespace={namespace} role={r} />
                ))}
              </ul>
            </>
          ) : (
            <>
              <p>The following error was returned:</p>
              <div className="error__content">
                <section className="Terminal terminal__error elevation-1 type-color-white error__text">
                  {rawMessage}
                </section>
              </div>
            </>
          )}
          <p>
            See the documentation for more info on{" "}
            <a
              href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/access-control.md"
              target="_blank"
              rel="noopener noreferrer"
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
