// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount, shallow } from "enzyme";
import {
  InstalledPackageDetail,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { has } from "lodash";
import { IK8sList, IKubeItem, IResource } from "shared/types";
import ApplicationStatus from "./ApplicationStatus";

const defaultProps = {
  deployments: [],
  statefulsets: [],
  daemonsets: [],
};

it("renders a loading status", () => {
  const deployments = [
    {
      isFetching: true,
    },
  ];
  const wrapper = shallow(<ApplicationStatus {...defaultProps} deployments={deployments} />);
  expect(wrapper.text()).toContain("Loading");
});

it("renders a deleting status", () => {
  const deployments = [
    {
      isFetching: false,
    },
  ];
  const wrapper = shallow(
    <ApplicationStatus
      {...defaultProps}
      deployments={deployments}
      info={
        {
          status: {
            ready: false,
            reason: InstalledPackageStatus_StatusReason.STATUS_REASON_UNINSTALLED,
            userReason: "Deleted",
          } as InstalledPackageStatus,
        } as InstalledPackageDetail
      }
    />,
  );
  expect(wrapper.text()).toContain("Deleted");
});

it("renders a failed status", () => {
  const deployments = [
    {
      isFetching: false,
    },
  ];
  const wrapper = shallow(
    <ApplicationStatus
      {...defaultProps}
      deployments={deployments}
      info={
        {
          status: {
            ready: false,
            reason: InstalledPackageStatus_StatusReason.STATUS_REASON_FAILED,
            userReason: "Failed",
          } as InstalledPackageStatus,
        } as InstalledPackageDetail
      }
    />,
  );
  expect(wrapper.text()).toContain("Failed");
});

describe("isFetching", () => {
  const tests: Array<{
    title: string;
    deployments: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
    statefulsets: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
    daemonsets: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>;
    deployed: boolean;
    totalPods: number;
    readyPods: number;
    infoReason: InstalledPackageStatus_StatusReason;
  }> = [
    {
      title: "shows a warning if no workloads are present",
      deployments: [],
      statefulsets: [],
      daemonsets: [],
      deployed: false,
      totalPods: 0,
      readyPods: 0,
      infoReason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
    },
    {
      title: "shows a deploying status if there is a non deployed deployment",
      deployments: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
            spec: {
              replicas: 1,
            },
            status: {
              availableReplicas: 0,
            },
          } as IResource,
        },
      ],
      statefulsets: [],
      daemonsets: [],
      deployed: false,
      totalPods: 1,
      readyPods: 0,
      infoReason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
    },
    {
      title: "shows a deploying status if there is a non deployed statefulset",
      statefulsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
            spec: {
              replicas: 1,
            },
            status: {
              readyReplicas: 0,
            },
          } as IResource,
        },
      ],
      deployments: [],
      daemonsets: [],
      deployed: false,
      totalPods: 1,
      readyPods: 0,
      infoReason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
    },
    {
      title: "shows a deploying status if there is a non deployed daemonset",
      daemonsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
            status: {
              currentNumberScheduled: 1,
              numberReady: 0,
            },
          } as IResource,
        },
      ],
      deployments: [],
      statefulsets: [],
      deployed: false,
      totalPods: 1,
      readyPods: 0,
      infoReason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
    },
    {
      title: "shows a deployed status if it has a daemonset, deployment and statefulset deployed",
      daemonsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
            status: {
              currentNumberScheduled: 1,
              numberReady: 1,
            },
          } as IResource,
        },
      ],
      deployments: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
            spec: {
              replicas: 1,
            },
            status: {
              availableReplicas: 1,
            },
          } as IResource,
        },
      ],
      statefulsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
            spec: {
              replicas: 1,
            },
            status: {
              readyReplicas: 1,
            },
          } as IResource,
        },
      ],
      deployed: true,
      totalPods: 3,
      readyPods: 3,
      infoReason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
    },
    {
      title:
        "shows a deploying status if it has a daemonset, deployment (deployed) and statefulset (not deployed)",
      daemonsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo-ds" },
            status: {
              currentNumberScheduled: 1,
              numberReady: 1,
            },
          } as IResource,
        },
      ],
      deployments: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo-dp" },
            spec: {
              replicas: 1,
            },
            status: {
              availableReplicas: 1,
            },
          } as IResource,
        },
      ],
      statefulsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo-ss" },
            spec: {
              replicas: 1,
            },
            status: {
              readyReplicas: 0,
            },
          } as IResource,
        },
      ],
      deployed: true,
      totalPods: 3,
      readyPods: 2,
      infoReason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
    },
    {
      title:
        "shows a deploying status if it has a daemonset, deployment (deployed) and statefulset (not deployed) with lists",
      daemonsets: [
        {
          isFetching: false,
          item: {
            items: [
              {
                metadata: { name: "foo-ds" },
                status: {
                  currentNumberScheduled: 1,
                  numberReady: 1,
                },
              } as IResource,
            ],
          } as IK8sList<IResource, {}>,
        },
      ],
      deployments: [
        {
          isFetching: false,
          item: {
            items: [
              {
                metadata: { name: "foo-dp" },
                spec: {
                  replicas: 1,
                },
                status: {
                  availableReplicas: 1,
                },
              } as IResource,
            ],
          } as IK8sList<IResource, {}>,
        },
      ],
      statefulsets: [
        {
          isFetching: false,
          item: {
            items: [
              {
                metadata: { name: "foo-ss" },
                spec: {
                  replicas: 1,
                },
                status: {
                  readyReplicas: 0,
                },
              } as IResource,
            ],
          } as IK8sList<IResource, {}>,
        },
      ],
      deployed: true,
      totalPods: 3,
      readyPods: 2,
      infoReason: InstalledPackageStatus_StatusReason.STATUS_REASON_PENDING,
    },
  ];
  tests.forEach(t => {
    it(t.title, () => {
      const wrapper = mount(<ApplicationStatus {...defaultProps} />);
      wrapper.setProps({
        deployments: t.deployments,
        statefulsets: t.statefulsets,
        daemonsets: t.daemonsets,
        info: {
          status: {
            reason: t.infoReason,
          } as InstalledPackageStatus,
        } as InstalledPackageDetail,
      });
      wrapper.update();
      const getItem = (i?: IResource | IK8sList<IResource, {}>): IResource => {
        return has(i, "items") ? (i as IK8sList<IResource, {}>).items[0] : (i as IResource);
      };
      if (!t.deployments.length && !t.statefulsets.length && !t.daemonsets.length) {
        expect(wrapper.text()).toContain("No Workload Found");
        return;
      }
      expect(wrapper.text()).toContain(t.deployed ? "Ready" : "Not Ready");
      // expect(wrapper.state()).toMatchObject({ totalPods: t.totalPods, readyPods: t.readyPods });
      // Check tooltip text
      const tooltipText = wrapper.html();
      t.deployments.forEach(d => {
        const item = getItem(d.item);
        expect(tooltipText).toContain(
          `<td>${item.metadata.name}</td><td>${item.status.availableReplicas}/${item.spec.replicas}</td>`,
        );
      });
      t.statefulsets.forEach(d => {
        const item = getItem(d.item);
        expect(tooltipText).toContain(
          `<td>${item.metadata.name}</td><td>${item.status.readyReplicas}/${item.spec.replicas}</td>`,
        );
      });
      t.daemonsets.forEach(d => {
        const item = getItem(d.item);
        expect(tooltipText).toContain(
          `<td>${item.metadata.name}</td><td>${item.status.numberReady}/${item.status.currentNumberScheduled}</td>`,
        );
      });
    });
  });
});
