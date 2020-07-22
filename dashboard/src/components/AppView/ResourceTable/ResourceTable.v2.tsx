import React, { useEffect, useMemo } from "react";
import { IK8sList, IKubeItem, IResource, ISecret } from "shared/types";

import { CdsIcon } from "components/Clarity/clarity";
import Table from "components/js/Table";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper.v2";
import ResourceRef from "shared/ResourceRef";
import { DaemonSetColumns } from "./ResourceData/DaemonSet";
import { DeploymentColumns } from "./ResourceData/Deployment";
import { OtherResourceColumns } from "./ResourceData/OtherResource";
import { SecretColumns } from "./ResourceData/Secret";
import { ServiceColumns } from "./ResourceData/Service";
import { StatefulSetColumns } from "./ResourceData/StatefulSet";

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

function getColumns(r: ResourceRef) {
  switch (r.kind) {
    case "Deployment":
      return DeploymentColumns;
    case "StatefulSet":
      return StatefulSetColumns;
    case "DaemonSet":
      return DaemonSetColumns;
    case "Service":
      return ServiceColumns;
    case "Secret":
      return SecretColumns;
    default:
      return OtherResourceColumns;
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
    const columns = useMemo(() => getColumns(resourceRefs[0]), resourceRefs);
    // TODO(andresmgot): Add support for lists
    const resourcesWithoutLists = resourceRefs.filter(ref => {
      const resource = resources[ref.getResourceURL()];
      if (resource && resource.item && (resource.item as IK8sList<IResource, {}>).items) {
        return false;
      }
      return true;
    });
    const data = useMemo(
      () =>
        resourcesWithoutLists.map(ref =>
          getData(
            ref.name,
            columns.map(c => c.accessor),
            columns.map(c => c.getter),
            resources[ref.getResourceURL()] as IKubeItem<IResource>,
          ),
        ),
      [resourceRefs, resources],
    );
    section = (
      <section aria-labelledby={`${id}-table`}>
        {title && <h6 id={`${id}-table`}>{title}</h6>}
        <Table columns={columns} data={data} />
      </section>
    );
  }
  return section;
}

export default ResourceTable;
