import { flatten } from "lodash";
import { get } from "lodash";
import * as React from "react";
import PieChart from "react-minimal-pie-chart";
import ReactTooltip from "react-tooltip";

import { AlertTriangle } from "react-feather";
import isSomeResourceLoading from "../../components/AppView/helpers";
import { hapi } from "../../shared/hapi/release";
import { IK8sList, IKubeItem, IResource } from "../../shared/types";
import "./ApplicationStatus.css";

export interface IApplicationStatusProps {
  deployments: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
  statefulsets: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
  daemonsets: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
  info?: hapi.release.IInfo;
  watchWorkloads: () => void;
  closeWatches: () => void;
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
      [
        {
          workloads: this.flattenItemList(deployments),
          readyKey: "status.availableReplicas",
          totalKey: "spec.replicas",
        },
        {
          workloads: this.flattenItemList(statefulsets),
          readyKey: "status.readyReplicas",
          totalKey: "spec.replicas",
        },
        {
          workloads: this.flattenItemList(daemonsets),
          readyKey: "status.numberReady",
          totalKey: "status.currentNumberScheduled",
        },
      ].forEach(src => {
        src.workloads.forEach(w => {
          const wReady = get(w, src.readyKey, 0);
          const wTotal = get(w, src.totalKey, 0);
          if (wReady) {
            readyPods += wReady;
          }
          if (wTotal) {
            totalPods += wTotal;
          }
          workloads = workloads.concat({
            name: w.metadata.name,
            replicas: wTotal,
            readyReplicas: wReady,
          });
        });
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
    if (this.state.totalPods === 0) {
      return (
        <span className="ApplicationStatus ApplicationStatus--pending">No workload found</span>
      );
    }
    return (
      <div className="ApplicationStatusPieChart">
        <div data-tip={true} data-for="app-status">
          <h5 className="ApplicationStatusPieChart__title">{ready ? "Ready" : "Not Ready"}</h5>
          <PieChart
            data={[{ value: 1, color: `${ready ? "#1598CB" : "#F58220"}` }]}
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
          <div className="ApplicationStatusPieChart__label">
            <p className="ApplicationStatusPieChart__label__number">{readyPods}</p>
            <p className="ApplicationStatusPieChart__label__text">Pod{readyPods > 1 ? "s" : ""}</p>
          </div>
        </div>
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

  private flattenItemList(items: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>) {
    return flatten(
      items.map(i => {
        const itemList = i.item as IK8sList<IResource, {}>;
        if (itemList && itemList.items) {
          // If the item is a list, return the array of items
          return itemList.items;
        }
        return i.item as IResource;
      }),
      // Remove empty items
    ).filter(r => !!r);
  }
}

export default ApplicationStatus;
