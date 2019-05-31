import * as React from "react";

import { AlertTriangle } from "react-feather";
import isSomeResourceLoading from "../../components/AppView/helpers";
import Check from "../../icons/Check";
import Compass from "../../icons/Compass";
import { hapi } from "../../shared/hapi/release";
import {
  IDaemonsetStatus,
  IDeploymentStatus,
  IKubeItem,
  IResource,
  IStatefulsetStatus,
} from "../../shared/types";
import "./ApplicationStatus.css";

interface IApplicationStatusProps {
  deployments: Array<IKubeItem<IResource>>;
  statefulsets: Array<IKubeItem<IResource>>;
  daemonsets: Array<IKubeItem<IResource>>;
  info?: hapi.release.IInfo;
  watchDeployments: () => void;
  closeWatches: () => void;
}

class ApplicationStatus extends React.Component<IApplicationStatusProps> {
  public componentDidMount() {
    this.props.watchDeployments();
  }

  public componentWillUnmount() {
    this.props.closeWatches();
  }

  public render() {
    if (isSomeResourceLoading(this.props.deployments)) {
      return <span className="ApplicationStatus">Loading...</span>;
    }
    if (this.props.info && this.props.info.deleted) {
      return this.renderDeletedStatus();
    }
    return this.isReady() ? this.renderSuccessStatus() : this.renderPendingStatus();
  }

  private renderSuccessStatus() {
    return (
      <span className="ApplicationStatus ApplicationStatus--success">
        <Check className="icon padding-t-tiny" /> Deployed
      </span>
    );
  }

  private renderPendingStatus() {
    return (
      <span className="ApplicationStatus ApplicationStatus--pending">
        <Compass className="icon padding-t-tiny" /> Deploying
      </span>
    );
  }

  private areDeploymentsReady() {
    const { deployments } = this.props;
    return deployments.every(d => {
      if (d.item) {
        const status: IDeploymentStatus = d.item.status;
        return status.availableReplicas === status.replicas;
      }
      return false;
    });
  }

  private areStatefulsetsReady() {
    const { statefulsets } = this.props;
    return statefulsets.every(d => {
      if (d.item) {
        const status: IStatefulsetStatus = d.item.status;
        return status.readyReplicas === status.replicas;
      }
      return false;
    });
  }

  private areDaemonsetsReady() {
    const { daemonsets } = this.props;
    return daemonsets.every(d => {
      if (d.item) {
        const status: IDaemonsetStatus = d.item.status;
        return status.numberReady === status.currentNumberScheduled;
      }
      return false;
    });
  }

  private isReady() {
    return this.areDeploymentsReady() && this.areStatefulsetsReady() && this.areDaemonsetsReady();
  }

  private renderDeletedStatus() {
    return (
      <span className="ApplicationStatus ApplicationStatus--deleted">
        <AlertTriangle className="icon" style={{ bottom: "-0.425em" }} /> Deleted
      </span>
    );
  }
}

export default ApplicationStatus;
