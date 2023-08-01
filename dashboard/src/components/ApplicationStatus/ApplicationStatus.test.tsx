// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import {
  InstalledPackageDetail,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  ResourceRef,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { has } from "lodash";
import { Tooltip } from "react-tooltip";
import { initialKinds } from "reducers/kube";
import { keyForResourceRef } from "shared/ResourceRef";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IK8sList, IKubeItem, IKubeState, IResource } from "shared/types";
import ApplicationStatus from "./ApplicationStatus";

const defaultProps = {
  deployRefs: [],
  statefulsetRefs: [],
  daemonsetRefs: [],
};

it("renders a loading status", () => {
  const deployments = [
    {
      apiVersion: "v1",
      kind: "Deployment",
      namespace: "foo",
      name: "deployment-1",
    } as ResourceRef,
  ];
  const store = getStore({
    kube: {
      items: {
        "v1/Deployment/foo/deployment-1": {
          isFetching: true,
        },
      },
      kinds: initialKinds,
    },
  });
  const wrapper = mountWrapper(
    store,
    <ApplicationStatus {...defaultProps} deployRefs={deployments} />,
  );

  expect(wrapper.text()).toContain("Loading");
});

it("renders a deleting status", () => {
  const deployments = [
    {
      apiVersion: "v1",
      kind: "Deployment",
      namespace: "foo",
      name: "deployment-1",
    } as ResourceRef,
  ];
  const store = getStore({
    kube: {
      items: {
        "v1/Deployment/foo/deployment-1": {
          isFetching: false,
        },
      },
      kinds: initialKinds,
    },
  });
  const wrapper = mountWrapper(
    store,
    <ApplicationStatus
      {...defaultProps}
      deployRefs={deployments}
      info={
        {
          status: {
            ready: false,
            reason: InstalledPackageStatus_StatusReason.UNINSTALLED,
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
      apiVersion: "v1",
      kind: "Deployment",
      namespace: "foo",
      name: "deployment-1",
    } as ResourceRef,
  ];
  const store = getStore({
    kube: {
      items: {
        "v1/Deployment/foo/deployment-1": {
          isFetching: false,
        },
      },
      kinds: initialKinds,
    },
  });
  const wrapper = mountWrapper(
    store,
    <ApplicationStatus
      {...defaultProps}
      deployRefs={deployments}
      info={
        {
          status: {
            ready: false,
            reason: InstalledPackageStatus_StatusReason.FAILED,
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
      infoReason: InstalledPackageStatus_StatusReason.PENDING,
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
      infoReason: InstalledPackageStatus_StatusReason.PENDING,
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
      infoReason: InstalledPackageStatus_StatusReason.PENDING,
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
      infoReason: InstalledPackageStatus_StatusReason.PENDING,
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
      infoReason: InstalledPackageStatus_StatusReason.INSTALLED,
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
      infoReason: InstalledPackageStatus_StatusReason.PENDING,
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
      infoReason: InstalledPackageStatus_StatusReason.PENDING,
    },
  ];

  const getRefsAndStateForTest = (
    resources: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>,
    kind: string,
  ) => {
    const resourceRefs: ResourceRef[] = [];
    const kubeItems: IKubeState["items"] = {};
    resources.forEach(r => {
      const item = r.item;
      if (Array.isArray((item as IK8sList<IResource, {}>).items)) {
        (item as IK8sList<IResource, {}>).items.forEach(i => {
          const ref = new ResourceRef({
            apiVersion: "v1",
            kind: kind,
            namespace: "foo",
            name: i.metadata.name,
          });
          resourceRefs.push(ref);
          kubeItems[keyForResourceRef(ref)] = r;
        });
      } else {
        const ref = new ResourceRef({
          apiVersion: "v1",
          kind: kind,
          namespace: "foo",
          name: (item as IResource).metadata.name,
        });
        resourceRefs.push(ref);
        kubeItems[keyForResourceRef(ref)] = r;
      }
    });
    return { resourceRefs, kubeItems };
  };

  tests.forEach(t => {
    it(t.title, () => {
      // Calculate resource refs from input.
      const { resourceRefs: deployRefs, kubeItems: kubeItemsDeploys } = getRefsAndStateForTest(
        t.deployments,
        "Deployment",
      );
      const { resourceRefs: daemonsetRefs, kubeItems: kubeItemsDaemonSets } =
        getRefsAndStateForTest(t.daemonsets, "Daemonset");
      const { resourceRefs: statefulsetRefs, kubeItems: kubeItemsStatefulSets } =
        getRefsAndStateForTest(t.statefulsets, "Statefulset");

      const kubeState: IKubeState = {
        items: { ...kubeItemsDeploys, ...kubeItemsDaemonSets, ...kubeItemsStatefulSets },
        kinds: initialKinds,
      };

      const store = getStore({
        kube: kubeState,
      });

      const info = {
        status: {
          reason: t.infoReason,
        } as InstalledPackageStatus,
      } as InstalledPackageDetail;
      const wrapper = mountWrapper(
        store,
        <ApplicationStatus
          {...defaultProps}
          deployRefs={deployRefs}
          daemonsetRefs={daemonsetRefs}
          statefulsetRefs={statefulsetRefs}
          info={info}
        />,
      );

      if (!t.deployments.length && !t.statefulsets.length && !t.daemonsets.length) {
        expect(wrapper.text()).toContain("No Workload Found");
        return;
      }
      expect(wrapper.text()).toContain(t.deployed ? "Ready" : "Not Ready");

      // Cloning the tooltip with the isOpen prop set to true,
      // this way we can later test the tooltip content
      act(() => {
        wrapper.setProps({
          children: (
            <Tooltip {...wrapper.find(Tooltip).props()} isOpen={true}>
              {wrapper.find(Tooltip).prop("children")}
            </Tooltip>
          ),
        });
      });
      wrapper.update();

      expect(wrapper.find(Tooltip)).toExist();

      t.deployments.forEach(d => {
        const item = getItem(d.item);
        expect(wrapper.find(Tooltip).html()).toContain(
          `<td>${item.metadata.name}</td><td>${item.status.availableReplicas}/${item.spec.replicas}</td>`,
        );
      });
      t.statefulsets.forEach(d => {
        const item = getItem(d.item);
        expect(wrapper.find(Tooltip).html()).toContain(
          `<td>${item.metadata.name}</td><td>${item.status.readyReplicas}/${item.spec.replicas}</td>`,
        );
      });
      t.daemonsets.forEach(d => {
        const item = getItem(d.item);
        expect(wrapper.find(Tooltip).html()).toContain(
          `<td>${item.metadata.name}</td><td>${item.status.numberReady}/${item.status.currentNumberScheduled}</td>`,
        );
      });
    });
  });
});

function getItem(i?: IResource | IK8sList<IResource, {}>): IResource {
  return has(i, "items") ? (i as IK8sList<IResource, {}>).items[0] : (i as IResource);
}
