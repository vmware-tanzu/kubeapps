// NOTE!
// This component has been deprecated
// Please use UnexpectedErrorAlert instead
import * as React from "react";
import { X } from "react-feather";

import { definedNamespaces } from "../../shared/Namespace";
import ErrorPageHeader from "./ErrorAlertHeader";
import { namespaceText } from "./helpers";

interface INotFoundErrorPageProps {
  header?: string;
  children?: JSX.Element;
  resource?: string;
  namespace?: string;
}

class NotFoundErrorPage extends React.Component<INotFoundErrorPageProps> {
  public render() {
    const { children, header, resource, namespace } = this.props;
    return (
      <div className="alert alert-error margin-t-bigger">
        <ErrorPageHeader icon={X}>
          {header ? (
            header
          ) : (
            <span>
              {resource} not found
              {namespace && <span> in {namespaceText(namespace)}</span>}.
            </span>
          )}
        </ErrorPageHeader>
        {namespace === definedNamespaces.all && (
          <div className="error__content margin-l-enormous">
            <p>You may need to select a namespace.</p>
          </div>
        )}
        {children && <div className="error__content margin-l-enormous">{children}</div>}
      </div>
    );
  }
}

export default NotFoundErrorPage;
