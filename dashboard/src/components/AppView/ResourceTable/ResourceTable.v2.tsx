import React, { useEffect, useMemo } from "react";
import { IKubeItem, IResource, ISecret, IStoreState } from "shared/types";

import actions from "actions";
import { CdsIcon } from "components/Clarity/clarity";
import Table from "components/js/Table";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper.v2";
import { useDispatch, useSelector } from "react-redux";
import ResourceRef from "shared/ResourceRef";
import { flattenResources } from "shared/utils";
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

function ResourceTable({ id, title, resourceRefs }: IResourceTableProps) {
  const dispatch = useDispatch();
  useEffect(() => {
    resourceRefs.forEach(r => dispatch(actions.kube.getAndWatchResource(r)));
    return function cleanup() {
      resourceRefs.forEach(r => dispatch(actions.kube.closeWatchResource(r)));
    };
  }, [resourceRefs, dispatch]);
  const resources = useSelector((state: IStoreState) =>
    flattenResources(resourceRefs, state.kube.items),
  );

  const columns = useMemo(() => getColumns(resourceRefs[0]), [resourceRefs]);
  const data = useMemo(
    () =>
      resources.map((resource, index) =>
        getData(
          resourceRefs[index].name,
          columns.map(c => c.accessor),
          columns.map(c => c.getter),
          resource,
        ),
      ),
    [columns, resourceRefs, resources],
  );

  let section: JSX.Element | null = null;
  if (resourceRefs.length > 0) {
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
