// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import cs from "classnames";
import PropTypes from "prop-types";
import React from "react";
import TableRow from "./components/TableRow/TableRow";
import "./Table.scss";
import TableTypes from "./Table.types";

const Table = ({ className, columns, data, compact, noBorder, valign }) => {
  const cssClass = cs("table", className, {
    "table-compact": compact,
    "table-noborder": noBorder,
    [`table-valign-${valign}`]: valign !== "",
  });

  return (
    <table className={cssClass}>
      <thead>
        <tr>
          {columns.map(({ accessor, Header, align }) => {
            const css = cs({ [align]: align });
            return (
              <th key={accessor} className={css}>
                {Header}
              </th>
            );
          })}
        </tr>
      </thead>
      <tbody>
        {data && data.map((r, i) => <TableRow key={i} row={r} columns={columns} index={i} />)}
      </tbody>
    </table>
  );
};

Table.propTypes = {
  columns: TableTypes.columns,
  className: PropTypes.string,
  data: TableTypes.rows.isRequired,
  compact: PropTypes.bool,
  noBorder: PropTypes.bool,
  valign: PropTypes.string,
  vertical: PropTypes.bool,
};

Table.defaultProps = {
  className: "",
  compact: false,
  noBorder: false,
  valign: "",
  vertical: false,
};

export default Table;
