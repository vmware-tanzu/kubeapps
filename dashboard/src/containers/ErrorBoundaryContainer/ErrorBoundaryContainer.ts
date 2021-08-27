import { connect } from "react-redux";
import { IStoreState } from "shared/types";
import ErrorBoundary from "../../components/ErrorBoundary";

interface IErrorBoundaryProps {
  children: React.ReactChildren | React.ReactNode | string;
}

function mapStateToProps(
  { clusters: { currentCluster, clusters } }: IStoreState,
  { children }: IErrorBoundaryProps,
) {
  const cluster = clusters[currentCluster];
  return {
    error: cluster.error && cluster.error.error,
    children,
  };
}

export default connect(mapStateToProps)(ErrorBoundary);
