// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import isSomeResourceLoading from "components/AppView/helpers";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import {
  InstalledPackageDetail,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { flatten, get } from "lodash";
import { useEffect, useState } from "react";
import { PieChart } from "react-minimal-pie-chart";
import ReactTooltip from "react-tooltip";
import { IK8sList, IKubeItem, IResource } from "../../shared/types";
import "./ApplicationStatus.css";

interface IApplicationStatusProps {
  deployments: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
  statefulsets: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
  daemonsets: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
  info?: InstalledPackageDetail;
}

interface IWorkload {
  replicas: number;
  readyReplicas: number;
  name: string;
  type: string;
}

function flattenItemList(items: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>) {
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

function codeToString(status: InstalledPackageStatus | null | undefined) {
  const codes = {
    [InstalledPackageStatus_StatusReason.STATUS_REASON_UNSPECIFIED]: "Unknown",
    [InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED]: "Installed",
    [InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED]: "Deleted",
    [InstalledPackageStatus_StatusReason.STATUS_REASON_FAILED]: "Failed",
    [InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING]: "Pending",
    [InstalledPackageStatus_StatusReason.UNRECOGNIZED]: "Unknown",
  };
  let msg = codes[0];
  if (status && status.reason) {
    msg = codes[status.reason];
  }
  return `${msg}${status?.userReason ? " (" + status.userReason + ")" : ""}`;
}

export default function ApplicationStatus({
  deployments,
  statefulsets,
  daemonsets,
  info,
}: IApplicationStatusProps) {
  const [workloads, setWorkloads] = useState([] as IWorkload[]);
  const [totalPods, setTotalPods] = useState(0);
  const [readyPods, setReadyPods] = useState(0);

  useEffect(() => {
    let currentTotalPods = 0;
    let currentReadyPods = 0;
    let currentWorkloads: IWorkload[] = [];
    [
      {
        workloads: flattenItemList(deployments),
        readyKey: "status.availableReplicas",
        totalKey: "spec.replicas",
        type: "deployment",
      },
      {
        workloads: flattenItemList(statefulsets),
        readyKey: "status.readyReplicas",
        totalKey: "spec.replicas",
        type: "statefulset",
      },
      {
        workloads: flattenItemList(daemonsets),
        readyKey: "status.numberReady",
        totalKey: "status.currentNumberScheduled",
        type: "daemonset",
      },
    ].forEach(src => {
      src.workloads.forEach(w => {
        const wReady = get(w, src.readyKey, 0);
        const wTotal = get(w, src.totalKey, 0);
        if (wReady) {
          currentReadyPods += wReady;
        }
        if (wTotal) {
          currentTotalPods += wTotal;
        }
        currentWorkloads = currentWorkloads.concat({
          name: w.metadata.name,
          replicas: wTotal,
          readyReplicas: wReady,
          type: src.type,
        });
      });
    });
    setWorkloads(currentWorkloads);
    setReadyPods(currentReadyPods);
    setTotalPods(currentTotalPods);
  }, [deployments, statefulsets, daemonsets]);

  const ready = totalPods === readyPods;

  if (isSomeResourceLoading(deployments.concat(statefulsets).concat(daemonsets))) {
    return (
      <div className="status-loading-wrapper margin-t-xl">
        <LoadingWrapper loadingText="Loading..." size={"md"} />
      </div>
    );
  }
  if (info?.status?.reason === InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED) {
    return (
      <div className="center">
        <div className="color-icon-danger">
          <CdsIcon shape="exclamation-triangle" size="md" solid={true} /> Application Deleted
        </div>
      </div>
    );
  }
  if (info?.status?.reason) {
    // If the status code is different than "Deployed" or "Pending",
    // display that status.
    const packageStatus = codeToString(info.status);
    if (
      ![
        InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
        InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
      ].includes(info?.status?.reason)
    ) {
      return (
        <div className="center">
          <div className="color-icon-warning">
            <CdsIcon shape="exclamation-triangle" size="md" solid={true} /> Status: {packageStatus}
          </div>
        </div>
      );
    }
  }
  if (totalPods === 0) {
    return (
      <div className="center">
        <div className="color-icon-info">
          <CdsIcon shape="exclamation-triangle" size="md" solid={true} /> No Workload Found
        </div>
      </div>
    );
  }
  return (
    <section aria-label="Application status" className="application-status-pie-chart">
      <div data-tip={true} data-for="application-status">
        <h5 className="application-status-pie-chart-title">{ready ? "Ready" : "Not Ready"}</h5>
        <PieChart
          data={[{ value: 1, color: "#0072a3" }]}
          reveal={(readyPods / totalPods) * 100}
          animate={true}
          animationDuration={1000}
          lineWidth={20}
          startAngle={270}
          labelStyle={{ fontSize: "30px", fontWeight: 600 }}
          rounded={true}
          style={{ height: "100px", width: "100px" }}
          background="#bfbfbf"
        />
        <div className="application-status-pie-chart-label">
          <p className="application-status-pie-chart-number">{readyPods}</p>
          <p className="application-status-pie-chart-text">Pod{readyPods > 1 ? "s" : ""}</p>
        </div>
      </div>
      <ReactTooltip id="application-status" className="extraClass" effect="solid" place="right">
        <table className="application-status-table">
          <thead>
            <tr>
              <th>Workload</th>
              <th>Ready</th>
            </tr>
          </thead>
          <tbody>
            {workloads.map(workload => {
              return (
                <tr key={`${workload.type}/${workload.name}`}>
                  <td>{workload.name}</td>
                  <td>
                    {workload.readyReplicas}/{workload.replicas}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </ReactTooltip>
    </section>
  );
}
