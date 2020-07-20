import { get, isEmpty } from "lodash";
import React, { useEffect } from "react";
import {
  IK8sList,
  IKubeItem,
  IPort,
  IResource,
  ISecret,
  IServiceSpec,
  IServiceStatus,
} from "shared/types";

import { CdsIcon } from "components/Clarity/clarity";
import Table from "components/js/Table";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper.v2";
import ResourceRef from "shared/ResourceRef";
import SecretItemDatum from "./ResourceItem/SecretItem/SecretItemDatum.v2";

interface IResourceTableProps {
  id: string;
  title?: string;
  resourceRefs: ResourceRef[];
  resources: { [s: string]: IKubeItem<IResource | IK8sList<IResource, {}>> };
  requestResources?: boolean;
  watchResource: (r: ResourceRef) => void;
  closeWatch: (r: ResourceRef) => void;
  avoidEmptyResource?: boolean;
}

function getServiceExternalIP(service: IResource): string {
  const spec: IServiceSpec = service.spec;
  const status: IServiceStatus = service.status;
  if (spec.type !== "LoadBalancer") {
    return "None";
  }
  if (status.loadBalancer.ingress && status.loadBalancer.ingress.length > 0) {
    return (
      status.loadBalancer.ingress[0].hostname || status.loadBalancer.ingress[0].ip || "Pending"
    );
  }
  return "Pending";
}

function getServicePorts(service: IResource): string {
  if (service.spec.ports) {
    return service.spec.ports
      .map((p: IPort) => `${p.port}${p.nodePort ? `:${p.nodePort}` : ""}/${p.protocol || "TCP"}`)
      .join(", ");
  }
  return "";
}

function getSecretData(secret: ISecret) {
  const data = secret.data;
  if (isEmpty(data)) {
    return "This Secret is empty";
  }
  return Object.keys(data).map(k => (
    <SecretItemDatum key={`${secret.metadata.name}/${k}`} name={k} value={data[k]} />
  ));
}

function getColumns(r: ResourceRef) {
  switch (r.kind) {
    case "Deployment":
      return [
        {
          accessor: "name",
          Header: "NAME",
          getter: (target: IResource) => get(target, "metadata.name"),
        },
        {
          accessor: "desired",
          Header: "DESIRED",
          getter: (target: IResource) => get(target, "status.replicas"),
        },
        {
          accessor: "upToDate",
          Header: "UP-TO-DATE",
          getter: (target: IResource) => get(target, "status.updatedReplicas"),
        },
        {
          accessor: "available",
          Header: "AVAILABLE",
          getter: (target: IResource) => get(target, "status.availableReplicas"),
        },
      ];
    case "StatefulSet":
      return [
        {
          accessor: "name",
          Header: "NAME",
          getter: (target: IResource) => get(target, "metadata.name"),
        },
        {
          accessor: "desired",
          Header: "DESIRED",
          getter: (target: IResource) => get(target, "spec.replicas"),
        },
        {
          accessor: "upToDate",
          Header: "UP-TO-DATE",
          getter: (target: IResource) => get(target, "status.updatedReplicas"),
        },
        {
          accessor: "ready",
          Header: "READY",
          getter: (target: IResource) => get(target, "status.readyReplicas"),
        },
      ];
    case "DaemonSet":
      return [
        {
          accessor: "name",
          Header: "NAME",
          getter: (target: IResource) => get(target, "metadata.name"),
        },
        {
          accessor: "desired",
          Header: "DESIRED",
          getter: (target: IResource) => get(target, "status.currentNumberScheduled"),
        },
        {
          accessor: "available",
          Header: "AVAILABLE",
          getter: (target: IResource) => get(target, "status.numberReady"),
        },
      ];
    case "Service":
      return [
        {
          accessor: "name",
          Header: "NAME",
          getter: (target: IResource) => get(target, "metadata.name"),
        },
        {
          accessor: "type",
          Header: "TYPE",
          getter: (target: IResource) => get(target, "spec.type"),
        },
        {
          accessor: "clusterIP",
          Header: "CLUSTER-IP",
          getter: (target: IResource) => get(target, "spec.clusterIP"),
        },
        {
          accessor: "externalIP",
          Header: "EXTERNAL-IP",
          getter: (target: IResource) => getServiceExternalIP(target),
        },
        {
          accessor: "ports",
          Header: "PORT(S)",
          getter: (target: IResource) => getServicePorts(target),
        },
      ];
    case "Secret":
      return [
        {
          accessor: "name",
          Header: "NAME",
          getter: (target: IResource) => get(target, "metadata.name"),
        },
        {
          accessor: "type",
          Header: "TYPE",
          getter: (target: IResource) => get(target, "type"),
        },
        {
          accessor: "data",
          Header: "DATA",
          getter: (target: ISecret) => getSecretData(target),
        },
      ];
    default:
      return [
        {
          accessor: "name",
          Header: "NAME",
          getter: (target: IResource) => get(target, "metadata.name"),
        },
        {
          accessor: "kind",
          Header: "KIND",
          getter: (target: IResource) => get(target, "kind"),
        },
      ];
  }
}

function getData(
  name: string,
  accessors: string[],
  getters: Array<(r: any) => string | JSX.Element | JSX.Element[]>,
  resource?: IKubeItem<IResource | ISecret>,
) {
  const data = {
    name,
  };
  if (!resource || resource.isFetching) {
    data[accessors[1]] = <LoadingWrapper small={true} />;
    return data;
  }
  if (resource.error) {
    data[accessors[1]] = (
      <>
        <CdsIcon shape="alert-triangle" />
        Error: {resource.error.message}
      </>
    );
    return data;
  }
  if (resource.item) {
    accessors.forEach((accessor, index) => {
      data[accessor] = getters[index](resource.item);
    });
    return data;
  }
  data[accessors[1]] = <span>Unkown</span>;
  return;
}

function ResourceTable({
  id,
  title,
  resourceRefs,
  requestResources,
  resources,
  watchResource,
  closeWatch,
}: IResourceTableProps) {
  useEffect(() => {
    resourceRefs.forEach(r => watchResource(r));
    return function cleanup() {
      resourceRefs.forEach(r => closeWatch(r));
    };
  }, [requestResources, resourceRefs, watchResource, closeWatch]);

  let section: JSX.Element | null = null;
  if (resourceRefs.length > 0) {
    const columns = getColumns(resourceRefs[0]);
    // TODO(andresmgot): Add support for lists
    const resourcesWithoutLists = resourceRefs.filter(ref => {
      const resource = resources[ref.getResourceURL()];
      if (resource && resource.item && (resource.item as IK8sList<IResource, {}>).items) {
        return false;
      }
      return true;
    });
    section = (
      <section aria-labelledby={`${id}-table`}>
        {title && <h6 id={`${id}-table`}>{title}</h6>}
        <Table
          columns={columns}
          data={resourcesWithoutLists.map(ref =>
            getData(
              ref.name,
              columns.map(c => c.accessor),
              columns.map(c => c.getter),
              resources[ref.getResourceURL()] as IKubeItem<IResource>,
            ),
          )}
        />
      </section>
    );
  }
  return section;
}

export default ResourceTable;
