import * as React from "react";
import { AppConflict, ForbiddenError, IRBACRole, NotFoundError } from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "kubeapps.com",
    namespace: "kubeapps",
    resource: "apprepositories",
    verbs: ["get"],
  },
];

interface IDeploymentErrorProps {
  namespace: string;
  releaseName: string;
  error: Error | undefined;
}

class DeploymentErrors extends React.Component<IDeploymentErrorProps> {
  public render() {
    const { error, namespace, releaseName } = this.props;
    switch (error && error.constructor) {
      case AppConflict:
        return (
          <NotFoundErrorAlert
            header={`The given release name already exists in the cluster. Choose a different one`}
          />
        );
      case ForbiddenError:
        const message = error ? error.message : "";
        let roles: IRBACRole[] = [];
        try {
          roles = JSON.parse(message);
        } catch (e) {
          // Cannot parse the error as a role array
          // return the default roles
          roles = RequiredRBACRoles;
        }
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={roles}
            action={`deploy the application "${releaseName}"`}
          />
        );

      case NotFoundError:
        return (
          <NotFoundErrorAlert resource={`Application "${releaseName}"`} namespace={namespace} />
        );
      default:
        return <UnexpectedErrorAlert />;
    }
  }
}

export default DeploymentErrors;
