// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import cs from "classnames";
import PropTypes from "prop-types";
import React from "react";
import TableTypes from "../../Table.types";

const TableRow = ({ columns, row, index }) => (
  <tr>
    {columns.map(({ accessor, align }) => {
      // Prioritize the dataKey value of the columns. If not, use key as a fallback
      const data = row[accessor];
      const css = cs({ [align]: align });

      return (
        <td key={`${index}-${accessor}`} className={css}>
          {data || "-"}
        </td>
      );
    })}
  </tr>
);

TableRow.propTypes = {
  columns: TableTypes.columns,
  row: TableTypes.row,
  // It's required because we need it to check the `TableTypes.row` validator.
  index: PropTypes.number.isRequired,
};

export default TableRow;
