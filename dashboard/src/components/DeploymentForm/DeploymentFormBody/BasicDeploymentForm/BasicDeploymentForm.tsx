// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CellContext, ColumnDef, createColumnHelper } from "@tanstack/react-table";
import React, { useMemo, useState } from "react";
import { DeploymentEvent, IBasicFormParam } from "shared/types";
import "./BasicDeploymentForm.css";
import TabularSchemaEditorTable from "./TabularSchemaEditorTable/TabularSchemaEditorTable";
import { fuzzySort } from "./TabularSchemaEditorTable/TabularSchemaEditorTableHelpers";
import {
  renderConfigCurrentValuePro,
  renderConfigDefaultValue,
  renderConfigDeployedValue,
  renderConfigDescription,
  renderConfigKey,
  renderConfigKeyHeader,
  renderConfigType,
} from "./TabularSchemaEditorTable/TabularSchemaEditorTableRenderer";

export interface IBasicDeploymentFormProps {
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
  deploymentEvent: DeploymentEvent;
  paramsFromComponentState: IBasicFormParam[];
  isLoading: boolean;
  saveAllChanges: () => void;
}

function BasicDeploymentForm(props: IBasicDeploymentFormProps) {
  // Fetch data from the parent component
  const {
    handleBasicFormParamChange,
    saveAllChanges,
    deploymentEvent,
    paramsFromComponentState,
    isLoading,
  } = props;

  // Component state
  const [globalFilter, setGlobalFilter] = useState("");

  // Column definitions
  // use useMemo to avoid re-creating the columns on every render
  const columnHelper = createColumnHelper<IBasicFormParam>();
  const columns = useMemo<ColumnDef<IBasicFormParam>[]>(() => {
    const cols = [
      columnHelper.accessor((row: IBasicFormParam) => row.key, {
        id: "key",
        cell: (info: CellContext<IBasicFormParam, any>) =>
          renderConfigKey(info.row.original, info.row, saveAllChanges),
        header: info => renderConfigKeyHeader(info.table, saveAllChanges),
        sortingFn: fuzzySort,
      }),
      columnHelper.accessor((row: IBasicFormParam) => row.type, {
        id: "type",
        cell: (info: CellContext<IBasicFormParam, any>) =>
          renderConfigType(info.row.original, info.row),
        header: () => <span>Type</span>,
      }),
      columnHelper.accessor((row: IBasicFormParam) => row.description, {
        id: "description",
        cell: (info: CellContext<IBasicFormParam, any>) =>
          renderConfigDescription(info.row.original, info.row),
        header: () => <span>Description</span>,
      }),
      columnHelper.accessor((row: IBasicFormParam) => row.defaultValue, {
        id: "defaultValue",
        cell: (info: CellContext<IBasicFormParam, any>) =>
          renderConfigDefaultValue(info.row.original, info.row),
        header: () => <span>Default Value</span>,
      }),
      columnHelper.accessor((row: IBasicFormParam) => row.currentValue, {
        id: "currentValue",
        cell: (info: CellContext<IBasicFormParam, any>) => {
          return renderConfigCurrentValuePro(info.row.original, handleBasicFormParamChange);
        },
        header: () => <span>Current Value</span>,
      }),
    ];
    if (deploymentEvent === "upgrade") {
      cols.splice(
        4,
        0,
        columnHelper.accessor((row: IBasicFormParam) => row.deployedValue, {
          id: "deployedValue",
          cell: (info: CellContext<IBasicFormParam, any>) =>
            renderConfigDeployedValue(info.row.original, info.row),
          header: () => <span>Deployed Value</span>,
        }),
      );
    }
    return cols;
  }, [columnHelper, deploymentEvent, handleBasicFormParamChange, saveAllChanges]);

  return (
    <TabularSchemaEditorTable
      columns={columns}
      data={paramsFromComponentState}
      globalFilter={globalFilter}
      setGlobalFilter={setGlobalFilter}
      isLoading={isLoading}
      saveAllChanges={saveAllChanges}
    />
  );
}

export default BasicDeploymentForm;
