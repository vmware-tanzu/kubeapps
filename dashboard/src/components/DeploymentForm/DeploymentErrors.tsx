import * as React from "react";
import {
  AppConflict,
  ForbiddenError,
  IRBACRole,
  MissingChart,
  NotFoundError,
  UnprocessableEntity,
} from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";

interface IDeploymentErrorProps {
  kubeappsNamespace: string;
  namespace: string;
  releaseName: string;
  error: Error | undefined;
  chartName: string;
  repo: string;
  version: string;
}

class DeploymentErrors extends React.Component<IDeploymentErrorProps> {
  public render() {
    const { chartName, error, namespace, releaseName, repo, version } = this.props;
    switch (error && error.constructor) {
      case AppConflict:
        return (
          <NotFoundErrorAlert
            header={"The given release name already exists in the cluster. Choose a different one"}
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
          roles = this.requiredRBACRoles();
        }
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={roles}
            action={`deploy the application "${releaseName}"`}
          />
        );
      case MissingChart:
        return (
          <NotFoundErrorAlert header={`Chart ${chartName} (v${version}) not found in ${repo}`} />
        );
      case NotFoundError:
        return (
          <NotFoundErrorAlert resource={`Application "${releaseName}"`} namespace={namespace} />
        );
      case UnprocessableEntity:
        return <UnexpectedErrorAlert text={error && error.message} raw={true} />;
      default:
        return <UnexpectedErrorAlert />;
    }
  }

  private requiredRBACRoles(): IRBACRole[] {
    return [
      {
        apiGroup: "kubeapps.com",
        namespace: this.props.kubeappsNamespace,
        resource: "apprepositories",
        verbs: ["get"],
      },
    ];
  }
}

export default DeploymentErrors;
