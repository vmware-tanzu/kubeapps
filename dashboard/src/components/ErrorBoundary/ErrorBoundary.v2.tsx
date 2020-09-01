import Alert from "components/js/Alert";
import * as React from "react";

export interface IErrorBoundaryProps {
  error?: Error;
  children: React.ReactChildren | React.ReactNode | string;
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
    const { error: stateError } = this.state;
    const { error: propsError } = this.props;
    if (propsError) {
      return <Alert theme="danger">Found an error: {propsError.message}</Alert>;
    }
    if (stateError) {
      return <Alert theme="danger">Found an error: {stateError.message}</Alert>;
    }
    return this.props.children;
  }

  public componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    this.setState({ error, errorInfo });
  }
}

export default ErrorBoundary;
