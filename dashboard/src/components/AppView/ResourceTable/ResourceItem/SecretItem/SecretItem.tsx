import { isEmpty } from "lodash";
import * as React from "react";
import { ISecret } from "shared/types";
import "./SecretContent.css";
import SecretItemDatum from "./SecretItemDatum";

interface ISecretItemRow {
  resource: ISecret;
}

export const SecretColumns: React.SFC = () => {
  return (
    <React.Fragment>
      <th className="col-3">NAME</th>
      <th className="col-2">TYPE</th>
      <th className="col-7">DATA</th>
    </React.Fragment>
  );
};

const SecretItem: React.SFC<ISecretItemRow> = props => {
  const { resource } = props;
  return (
    <React.Fragment>
      <td className="col-3">{resource.metadata.name}</td>
      <td className="col-2 SecretType">{resource.type}</td>
      {isEmpty(resource.data) ? (
        <td className="col-7">
          <span>This Secret is empty</span>
        </td>
      ) : (
        <td className="col-7">
          {Object.keys(resource.data).map(k => (
            <SecretItemDatum
              key={`${resource.metadata.name}/${k}`}
              name={k}
              value={resource.data[k]}
            />
          ))}
        </td>
      )}
    </React.Fragment>
  );
};

export default SecretItem;
