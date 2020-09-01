import { connect } from "react-redux";

import ErrorBoundary from "../../components/ErrorBoundary";
import { IStoreState } from "../../shared/types";

interface IErrorBoundaryProps {
  children: React.ReactChildren | React.ReactNode | string;
}

function mapStateToProps(
  { clusters: { currentCluster, clusters }, config }: IStoreState,
  { children }: IErrorBoundaryProps,
) {
  const cluster = clusters[currentCluster];
  return {
    error: cluster.error && cluster.error.error,
    children,
    UI: config.featureFlags.ui,
  };
}

export default connect(mapStateToProps)(ErrorBoundary);
