import * as React from "react";
import { IRBACRole } from "shared/types";
import { definedNamespaces } from "../../shared/Namespace";

import { namespaceText } from "./helpers";

interface IPermissionsListItemProps {
  role: IRBACRole;
  namespace: string;
}

class PermissionsListItem extends React.Component<IPermissionsListItemProps> {
  public render() {
    const { role } = this.props;
    const namespace = role.clusterWide
      ? definedNamespaces.all
      : role.namespace || this.props.namespace;
    return (
      <li>
        {role.verbs.join(", ")} <code>{role.resource}</code>{" "}
        {role.apiGroup !== "" ? (
          <span>
            (<code>{role.apiGroup}</code>)
          </span>
        ) : (
          ""
        )}{" "}
        in {namespaceText(namespace)}.
      </li>
    );
  }
}

export default PermissionsListItem;
