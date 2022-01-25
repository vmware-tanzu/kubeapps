// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { get, isEmpty } from "lodash";
import { IResource, ISecret } from "shared/types";
import SecretItemDatum from "../ResourceItem/SecretItem/SecretItemDatum";
import "./Secret.css";

function getSecretData(secret: ISecret) {
  const data = secret.data;
  if (isEmpty(data)) {
    return "This Secret is empty";
  }
  return Object.keys(data).map(k => (
    <div key={`cel-${secret.metadata.name}/${k}`} className="secret-table-cel">
      <div className="secret-table-content">
        <SecretItemDatum key={`${secret.metadata.name}/${k}`} name={k} value={data[k]} />
      </div>
    </div>
  ));
}

export const SecretColumns = [
  {
    accessor: "name",
    Header: "NAME",
    getter: (target: IResource) => get(target, "metadata.name"),
  },
  {
    accessor: "type",
    Header: "TYPE",
    getter: (target: IResource) => get(target, "type"),
  },
  {
    accessor: "data",
    Header: "DATA",
    getter: (target: ISecret) => getSecretData(target),
  },
];
