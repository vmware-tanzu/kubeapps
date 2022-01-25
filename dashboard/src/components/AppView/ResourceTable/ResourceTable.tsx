// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import Table from "components/js/Table";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import { useMemo } from "react";
import { useSelector } from "react-redux";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { IKubeItem, IResource, ISecret, IStoreState } from "shared/types";
import { DaemonSetColumns } from "./ResourceData/DaemonSet";
import { DeploymentColumns } from "./ResourceData/Deployment";
import { OtherResourceColumns } from "./ResourceData/OtherResource";
import { SecretColumns } from "./ResourceData/Secret";
import { ServiceColumns } from "./ResourceData/Service";
import { StatefulSetColumns } from "./ResourceData/StatefulSet";
import { filterByResourceRefs } from "containers/helpers";

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
    data[accessors[1]] = <LoadingWrapper size={"sm"} />;
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
  data[accessors[1]] = <span>Unknown</span>;
  return;
}

function ResourceTable({ id, title, resourceRefs }: IResourceTableProps) {
  const resources = useSelector((state: IStoreState) =>
    filterByResourceRefs(resourceRefs, state.kube.items),
  ) as IKubeItem<IResource>[];

  const columns = useMemo(
    () => (resourceRefs.length ? getColumns(resourceRefs[0]) : OtherResourceColumns),
    [resourceRefs],
  );
  const data = useMemo(
    () =>
      resources.map((resource, index) => {
        // When the resourceRef is a list, the list of references will be just one
        const ref = resourceRefs.length === 1 ? resourceRefs[0] : resourceRefs[index];
        if (ref) {
          return getData(
            ref.name,
            columns.map(c => c.accessor),
            columns.map(c => c.getter),
            resource,
          );
        }
        return {};
      }),
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
