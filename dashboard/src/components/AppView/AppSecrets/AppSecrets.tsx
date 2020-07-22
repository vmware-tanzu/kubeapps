import { isEmpty } from "lodash";
import React from "react";
import { useSelector } from "react-redux";
import { ISecret, IStoreState } from "shared/types";

import ResourceRef from "shared/ResourceRef";
import { flattenResources } from "shared/utils";
import SecretItemDatum from "../ResourceTable/ResourceItem/SecretItem/SecretItemDatum.v2";
import "./AppSecrets.css";

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
    flattenResources(secretRefs, state.kube.items),
  );
  let content;
  if (secretRefs.length === 0) {
    content = "The current application does not include secrets";
  } else {
    content = secrets.map(secret => {
      if (secret && secret.item) {
        const secretItem = secret.item as ISecret;
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
