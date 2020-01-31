import { connect } from "react-redux";

import ErrorBoundary from "../../components/ErrorBoundary";
import { IStoreState } from "../../shared/types";

interface IErrorBoundaryProps {
  children: React.ReactChildren | React.ReactNode | string;
}

function mapStateToProps({ namespace }: IStoreState, { children }: IErrorBoundaryProps) {
  return {
    error: namespace.error && namespace.error.error,
    children,
  };
}

export default connect(mapStateToProps)(ErrorBoundary);
