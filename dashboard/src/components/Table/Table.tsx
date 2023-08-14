// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import "./Table.css";
import TableRow from "./components/TableRow/TableRow";

export interface ITableProps {
  className?: string;
  columns: ITableColumnProps[];
  compact?: boolean;
  data: { [key: string]: any };
  noBorder?: boolean;
  valign?: string;
}

export interface ITableColumnProps {
  Header: string;
  accessor: string;
  align?: string;
  title?: string;
}

const Table = ({
  className = "",
  columns,
  compact = false,
  data,
  noBorder = false,
  valign = "",
}: ITableProps) => {
  let cssClass = "table";
  if (className) {
    cssClass = `${cssClass} ${className}`;
  }
  if (compact) {
    cssClass = `${cssClass} table-compact`;
  }
  if (noBorder) {
    cssClass = `${cssClass} table-noborder`;
  }
  if (valign) {
    cssClass = `${cssClass} table-valign-${valign}`;
  }

  return (
    <table className={cssClass}>
      <thead>
        <tr>
          {columns.map(({ accessor, Header, align }) => {
            return (
              <th key={accessor} className={align ? "align" : ""}>
                {Header}
              </th>
            );
          })}
        </tr>
      </thead>
      <tbody>
        {data &&
          data.map((r: { [key: string]: any }, i: number) => (
            <TableRow key={i} row={r} columns={columns} index={i} />
          ))}
      </tbody>
    </table>
  );
};

export default Table;
