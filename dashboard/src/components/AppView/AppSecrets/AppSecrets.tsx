// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import LoadingWrapper from "components/LoadingWrapper";
import { isEmpty } from "lodash";
import { useSelector } from "react-redux";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { ISecret, IStoreState } from "shared/types";
import SecretItemDatum from "../ResourceTable/ResourceItem/SecretItem/SecretItemDatum";
import "./AppSecrets.css";
import { filterByResourceRefs } from "containers/helpers";

interface IResourceTableProps {
  secretRefs: ResourceRef[];
}

function getSecretData(secret: ISecret) {
  const data = secret.data;
  if (isEmpty(data)) {
    return null;
  }
  return Object.keys(data).map(k => (
    <div key={`${secret.metadata.name}/${k}`}>
      <SecretItemDatum name={k} value={data[k]} />
    </div>
  ));
}

function AppSecrets({ secretRefs }: IResourceTableProps) {
  const secrets = useSelector((state: IStoreState) =>
    filterByResourceRefs(secretRefs, state.kube.items),
  );
  let content;
  if (secrets.some(s => s.isFetching)) {
    content = <LoadingWrapper />;
  } else if (secretRefs.length === 0) {
    content = "The current application does not include secrets";
  } else {
    content = secrets.map(secret => {
      if (secret && secret.item) {
        const secretItem = secret.item as unknown as ISecret;
        return getSecretData(secretItem);
      }
      return null;
    });
  }
  return (
    <section aria-labelledby="app-secrets">
      <h5 className="section-title" id="app-secrets">
        Application Secrets
      </h5>
      <div className="app-secrets-content">{content}</div>
    </section>
  );
}

export default AppSecrets;
