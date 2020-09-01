import * as React from "react";

import UnexpectedErrorPage from "../../components/ErrorAlert/UnexpectedErrorAlert";
import { UnexpectedErrorAlert } from "../ErrorAlert";

export interface IErrorBoundaryProps {
  error?: Error;
  children: React.ReactChildren | React.ReactNode | string;
}

interface IErrorBoundaryState {
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
}

class ErrorBoundary extends React.Component<IErrorBoundaryProps, IErrorBoundaryState> {
  public state: IErrorBoundaryState = { error: null, errorInfo: null };

  public render() {
    const { error } = this.state;
    if (this.props.error) {
      return (
        <UnexpectedErrorPage
          raw={true}
          showGenericMessage={false}
          text={this.props.error.message}
        />
      );
    }
    return <React.Fragment>{error ? this.renderError() : this.props.children}</React.Fragment>;
  }

  public componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    this.setState({ error, errorInfo });
  }

  private renderError() {
    return <UnexpectedErrorAlert showGenericMessage={true} />;
  }
}

export default ErrorBoundary;
