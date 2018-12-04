import * as React from "react";
import { definedNamespaces } from "../../shared/Namespace";
import {
  ConflictError,
  ForbiddenError,
  IRBACRole,
  NotFoundError,
  UnprocessableEntity,
} from "../../shared/types";
import { PermissionsErrorAlert, UnexpectedErrorAlert } from "./";
import { namespaceText } from "./helpers";

interface IErrorSelectorProps {
  error: Error;
  // Resource causing the failure (e.g. "Application my-wordpress")
  resource: string;
  // Action related to the error (e.g. create, delete)
  action?: string;
  // Default permissions that the component should have
  // returned in case it's a ForbiddenError and the message
  // doesn't contain specific RBAC roles
  defaultRequiredRBACRoles?: {
    [action: string]: IRBACRole[];
  };
  // Namespace in case the error is namespaced
  namespace?: string;
}

const ErrorSelector: React.SFC<IErrorSelectorProps> = props => {
  const { namespace, error, defaultRequiredRBACRoles, action, resource } = props;
  switch (error.constructor) {
    case ConflictError:
      return (
        <UnexpectedErrorAlert
          showGenericMessage={false}
          title={`${resource} already exists, try a different name.`}
        />
      );
    case ForbiddenError:
      const message = error ? error.message : "";
      let roles: IRBACRole[] = [];
      try {
        roles = JSON.parse(message);
        // Cannot parse the error as a role array
        // return the default roles
      } catch (e) {
        if (defaultRequiredRBACRoles && action) {
          roles = defaultRequiredRBACRoles[action];
        }
      }
      return (
        <PermissionsErrorAlert
          namespace={namespace || ""}
          roles={roles}
          action={`${action} ${resource || ""}`}
        />
      );
    case NotFoundError:
      let title: any;
      const titleText = `${resource} not found`;
      if (namespace) {
        title = (
          <span>
            {titleText} <span> in {namespaceText(namespace)}</span>{" "}
          </span>
        );
      } else {
        title = titleText;
      }
      return (
        <UnexpectedErrorAlert title={title} showGenericMessage={false}>
          {namespace === definedNamespaces.all && (
            <div className="error__content margin-l-enormous">
              <p>You may need to select a namespace.</p>
            </div>
          )}
        </UnexpectedErrorAlert>
      );
    case UnprocessableEntity:
      return (
        <UnexpectedErrorAlert
          title={`Sorry! Something went wrong processing ${resource}`}
          text={error.message}
          raw={true}
          showGenericMessage={false}
        />
      );
    default:
      return <UnexpectedErrorAlert raw={true} text={error.message} showGenericMessage={true} />;
  }
};

export default ErrorSelector;
