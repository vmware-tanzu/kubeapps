// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { ITableColumnProps } from "components/Table/Table";

export interface ITableRowProps {
  columns: ITableColumnProps[];
  row: { [a: string]: any };
  index: number;
}

const TableRow = ({ columns, row, index }: ITableRowProps) => (
  <tr>
    {columns.map(({ accessor, align }) => {
      // Prioritize the dataKey value of the columns. If not, use key as a fallback
      const data = row[accessor];

      return (
        <td key={`${index}-${accessor}`} className={align ? "align" : ""}>
          {data || "-"}
        </td>
      );
    })}
  </tr>
);

export default TableRow;
