// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import Alert from "components/js/Alert";
import React from "react";

export interface IErrorBoundaryProps {
  error?: Error;
  children: React.ReactChildren | React.ReactNode | string;
  logout: () => void;
}

interface IErrorBoundaryState {
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

// TODO(andresmgot): This component cannot be migrated to React Hooks (yet) because
// the hook for `componentDidCatch` has no equivalent.
// https://reactjs.org/docs/hooks-faq.html#do-hooks-cover-all-use-cases-for-classes
class ErrorBoundary extends React.Component<IErrorBoundaryProps, IErrorBoundaryState> {
  public state: IErrorBoundaryState = { error: null, errorInfo: null };

  public render() {
    // Errors in the state are caused by uncaught errors in children components
    // Errors in the props are related to namespace handling. In this case, show a logout
    // button in case the user wants to login with a privileged account.
    const { error: stateError } = this.state;
    const { error: propsError } = this.props;
    const err = propsError || stateError;
    if (err) {
      return (
        <Alert theme="danger">
          An error occurred: {err.message}.{" "}
          {propsError && (
            <CdsButton size="sm" action="flat" onClick={this.props.logout} type="button">
              Log out
            </CdsButton>
          )}
        </Alert>
      );
    }

    return this.props.children;
  }

  public componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    this.setState({ error, errorInfo });
  }
}

export default ErrorBoundary;
