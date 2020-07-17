import React from "react";
import PropTypes from "prop-types";
import TableTypes from "../../Table.types";
import cs from "classnames";

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
