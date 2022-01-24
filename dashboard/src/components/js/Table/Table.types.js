// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import PropTypes from "prop-types";

const columns = PropTypes.arrayOf(
  PropTypes.shape({
    accessor: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
    Header: PropTypes.string.isRequired,
    align: PropTypes.string,
  }),
).isRequired;

const row = PropTypes.object;

const rows = PropTypes.arrayOf(row);

export default {
  columns,
  row,
  rows,
};
