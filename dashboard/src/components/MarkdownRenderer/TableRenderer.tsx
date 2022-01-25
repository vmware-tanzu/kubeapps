// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

const TableRenderer: React.FunctionComponent<{}> = (props: any) => {
  return <table className="table">{props.children}</table>;
};

export default TableRenderer;
