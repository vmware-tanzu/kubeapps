import * as React from "react";
import PieChart from "react-minimal-pie-chart";
import * as ReactTooltip from "react-tooltip";

import { AlertTriangle } from "react-feather";
import isSomeResourceLoading from "../../components/AppView/helpers";
// import Check from "../../icons/Check";
// import Compass from "../../icons/Compass";
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
  watchWorkloads: () => void;
  closeWatches: () => void;
  skipPieChart?: boolean;
}

interface IWorkload {
  replicas: number;
  readyReplicas: number;
  name: string;
}

interface IApplicationStatusState {
  workloads: IWorkload[];
  totalPods: number;
  readyPods: number;
}

class ApplicationStatus extends React.Component<IApplicationStatusProps, IApplicationStatusState> {
  public state: IApplicationStatusState = {
    workloads: [],
    totalPods: 0,
    readyPods: 0,
  };

  public componentDidMount() {
    this.props.watchWorkloads();
  }

  public componentWillUnmount() {
    this.props.closeWatches();
  }

  public componentDidUpdate(prevProps: IApplicationStatusProps) {
    if (prevProps !== this.props) {
      const { deployments, statefulsets, daemonsets } = this.props;
      let totalPods = 0;
      let readyPods = 0;
      let workloads: IWorkload[] = [];
      deployments.forEach(d => {
        if (d.item) {
          const status: IDeploymentStatus = d.item.status;
          if (status.availableReplicas) {
            readyPods += status.availableReplicas;
          }
          if (status.replicas) {
            totalPods += status.replicas;
          }
          workloads = workloads.concat({
            name: d.item.metadata.name,
            replicas: status.replicas || 0,
            readyReplicas: status.availableReplicas || 0,
          });
        }
      });
      statefulsets.forEach(d => {
        if (d.item) {
          const status: IStatefulsetStatus = d.item.status;
          if (status.readyReplicas) {
            readyPods += status.readyReplicas;
          }
          if (status.replicas) {
            totalPods += status.replicas;
          }
          workloads = workloads.concat({
            name: d.item.metadata.name,
            replicas: status.replicas || 0,
            readyReplicas: status.readyReplicas || 0,
          });
        }
      });
      daemonsets.forEach(d => {
        if (d.item) {
          const status: IDaemonsetStatus = d.item.status;
          if (status.numberReady) {
            readyPods += status.numberReady;
          }
          if (status.currentNumberScheduled) {
            totalPods += status.currentNumberScheduled;
          }
          workloads = workloads.concat({
            name: d.item.metadata.name,
            replicas: status.currentNumberScheduled || 0,
            readyReplicas: status.numberReady || 0,
          });
        }
      });
      this.setState({ workloads, totalPods, readyPods });
    }
  }

  public render() {
    const { totalPods, readyPods } = this.state;
    const ready = totalPods === readyPods;

    if (isSomeResourceLoading(this.props.deployments)) {
      return <span className="ApplicationStatus">Loading...</span>;
    }
    if (this.props.info && this.props.info.deleted) {
      return this.renderDeletedStatus();
    }
    if (this.props.info && this.props.info.status) {
      // If the status code is different than "Deployed", display that status
      const helmStatus = this.codeToString(this.props.info.status);
      if (helmStatus !== "Deployed") {
        return this.helmStatusError(helmStatus);
      }
    }
    return (
      <div className="ApplicationStatusPieChart">
        <a data-tip={true} data-for="app-status">
          <h5 className="ApplicationStatusPieChart__title">{ready ? "Ready" : "Not Ready"}</h5>
          {/* Avoid issues when rendering the pie chart in tests */}
          {/* https://github.com/toomuchdesign/react-minimal-pie-chart/issues/131 */}
          {!this.props.skipPieChart && (
            <PieChart
              data={[{ value: 1, color: `${ready ? "#008145" : "#FDBA12"}` }]}
              reveal={(readyPods / totalPods) * 100}
              animate={true}
              animationDuration={1000}
              lineWidth={20}
              startAngle={270}
              labelStyle={{ fontSize: "30px" }}
              rounded={true}
              style={{ height: "100px", width: "100px" }}
              background="#bfbfbf"
            />
          )}
          <div className="ApplicationStatusPieChart__label">
            <p className="ApplicationStatusPieChart__label__number">{readyPods}</p>
            <p className="ApplicationStatusPieChart__label__text">Pod{readyPods > 1 ? "s" : ""}</p>
          </div>
        </a>
        <ReactTooltip id="app-status" className="extraClass" effect="solid" place="right">
          {this.renderWorkloadTable()}
        </ReactTooltip>
      </div>
    );
  }

  private renderWorkloadTable = () => {
    const { workloads } = this.state;
    return (
      <table style={{ margin: 0 }}>
        <thead>
          <tr>
            <th>Pod(s)</th>
            <th>Workload</th>
          </tr>
        </thead>
        <tbody>
          {workloads.map(workload => {
            return (
              <tr key={workload.name}>
                <td>
                  {workload.readyReplicas}/{workload.replicas}
                </td>
                <td>{workload.name}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    );
  };

  private codeToString(status: hapi.release.IStatus | null | undefined) {
    // Codes from https://github.com/helm/helm/blob/268695813ba957821e53a784ac849aa3ca7f70a3/_proto/hapi/release/status.proto
    const codes = {
      0: "Unknown",
      1: "Deployed",
      2: "Deleted",
      3: "Superseded",
      4: "Failed",
      5: "Deleting",
      6: "Pending Install",
      7: "Pending Upgrade",
      8: "Pending Rollback",
    };
    if (status && status.code) {
      return codes[status.code];
    }
    return codes[0];
  }

  private helmStatusError(status: string) {
    return (
      <span className="ApplicationStatus ApplicationStatus--error">
        <AlertTriangle className="icon" style={{ bottom: "-0.425em", left: "-0.3em" }} />
        {status}
      </span>
    );
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
