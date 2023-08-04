// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsSelect } from "@cds/react/select";
import {
  ColumnFiltersState,
  ExpandedState,
  flexRender,
  getCoreRowModel,
  getExpandedRowModel,
  getFacetedMinMaxValues,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import Column from "components/Column";
import LoadingWrapper from "components/LoadingWrapper";
import Row from "components/Row";
import { useState } from "react";
import { IBasicFormParam } from "shared/types";
import DebouncedInput from "./DebouncedInput";
import "./TabularSchemaEditorTable.css";
import { fuzzyFilter } from "./TabularSchemaEditorTableHelpers";

export interface TabularSchemaEditorTableProps {
  columns: any;
  data: IBasicFormParam[];
  globalFilter: any;
  setGlobalFilter: any;
  isLoading: boolean;
  saveAllChanges: () => void;
}

export default function TabularSchemaEditorTable(props: TabularSchemaEditorTableProps) {
  const { columns, data, globalFilter, setGlobalFilter, isLoading, saveAllChanges } = props;

  // Component state
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [globalExpanded, setGlobalExpanded] = useState<ExpandedState>({});

  const table = useReactTable({
    data,
    columns,
    state: {
      columnFilters,
      globalFilter,
      expanded: globalExpanded,
    },
    autoResetPageIndex: false,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getFacetedMinMaxValues: getFacetedMinMaxValues(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getSubRows: (row: IBasicFormParam) => row.params,
    globalFilterFn: fuzzyFilter,
    filterFromLeafRows: true,
    onColumnFiltersChange: setColumnFilters,
    onExpandedChange: setGlobalExpanded,
    onGlobalFilterChange: setGlobalFilter,
    enableColumnResizing: true,
  });

  const paginationButtons = (
    <>
      <div className="pagination-buttons">
        <CdsButton
          title="First page"
          style={{ marginRight: "0.5em" }}
          action="solid"
          status="primary"
          type="button"
          size="sm"
          onClick={() => {
            saveAllChanges();
            table.setPageIndex(0);
          }}
          disabled={!table.getCanPreviousPage()}
        >
          {"<<"}
        </CdsButton>
        <CdsButton
          title="Previous Page"
          style={{ marginRight: "0.5em" }}
          action="solid"
          status="primary"
          type="button"
          size="sm"
          onClick={() => {
            saveAllChanges();
            table.previousPage();
          }}
          disabled={!table.getCanPreviousPage()}
        >
          {"<"}
        </CdsButton>

        <CdsButton
          style={{ marginRight: "0.5em" }}
          action="flat"
          status="neutral"
          type="button"
          size="sm"
        >
          <span>
            Page{" "}
            <strong>
              {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
            </strong>
          </span>
        </CdsButton>
        <CdsButton
          title="Next Page"
          style={{ marginRight: "0.5em" }}
          action="solid"
          status="primary"
          type="button"
          size="sm"
          onClick={() => {
            saveAllChanges();
            table.nextPage();
          }}
          disabled={!table.getCanNextPage()}
        >
          {">"}
        </CdsButton>
        <CdsButton
          title="Last Page"
          style={{ marginRight: "0.5em" }}
          action="solid"
          status="primary"
          type="button"
          size="sm"
          onClick={() => {
            saveAllChanges();
            table.setPageIndex(table.getPageCount() - 1);
          }}
          disabled={!table.getCanNextPage()}
        >
          {">>"}
        </CdsButton>
      </div>
    </>
  );

  const topButtons = (
    <>
      <div className="table-control">
        <Row>
          <Column span={8}>{paginationButtons}</Column>
          <Column span={4}>
            <DebouncedInput
              title="Search"
              value={globalFilter ?? ""}
              onChange={value => {
                saveAllChanges();
                setGlobalFilter(String(value));
              }}
              placeholder="Type to search by key..."
            />
          </Column>
        </Row>
      </div>
    </>
  );
  const bottomButtons = (
    <>
      <div className="table-control">
        <Row>
          <Column span={8}>{paginationButtons}</Column>
          <Column span={4}>
            <CdsSelect>
              <label htmlFor="page-size">Page size</label>
              <select
                id="page-size"
                value={table.getState().pagination.pageSize}
                onChange={e => {
                  saveAllChanges();
                  table.setPageSize(Number(e.target.value));
                }}
              >
                {[10, 20, 30, 40, 50].map(pageSize => (
                  <option key={pageSize} value={pageSize}>
                    Show {pageSize}
                  </option>
                ))}
              </select>
            </CdsSelect>
          </Column>
        </Row>
      </div>
    </>
  );

  const tableHeader = table.getHeaderGroups().map((headerGroup: any) => (
    <tr key={headerGroup.id}>
      {headerGroup.headers.map((header: any) => (
        <th key={header.id}>
          {header.isPlaceholder
            ? null
            : flexRender(header.column.columnDef.header, header.getContext())}
        </th>
      ))}
    </tr>
  ));

  const tableBody = table.getRowModel().rows.map((row: any) => (
    <tr key={row.id}>
      {row.getVisibleCells().map((cell: any) => (
        <td key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>
      ))}
    </tr>
  ));

  const tableFooter = table.getFooterGroups().map((footerGroup: any) => (
    <tr key={footerGroup.id}>
      {footerGroup.headers.map((header: any) => (
        <th key={header.id}>
          {header.isPlaceholder
            ? null
            : flexRender(header.column.columnDef.footer, header.getContext())}
        </th>
      ))}
    </tr>
  ));

  const tableObject = (
    <table className="table table-valign-center">
      <thead>{tableHeader}</thead>
      <tbody>{tableBody}</tbody>
      <tfoot>{tableFooter}</tfoot>
    </table>
  );

  return (
    <LoadingWrapper loaded={!isLoading}>
      {topButtons}
      {tableObject}
      {bottomButtons}
    </LoadingWrapper>
  );
}
