import * as React from "react";

import "./AccessURLItem.css";
import { IURLItem } from "./IURLItem";

interface IAccessURLItem {
  URLItem: IURLItem;
}

const AccessURLItem: React.SFC<IAccessURLItem> = props => {
  const { URLItem } = props;
  return (
    <tr>
      <td>{URLItem.name}</td>
      <td>{URLItem.type}</td>
      <td>
        {URLItem.URLs.map(l => (
          <span
            key={l}
            className={`ServiceItem ${
              URLItem.isLink ? "ServiceItem--with-link" : ""
            } type-small margin-r-small padding-tiny padding-h-normal`}
          >
            {URLItem.isLink ? (
              <a className="padding-tiny padding-h-normal" href={l} target="_blank">
                {l}
              </a>
            ) : (
              l
            )}
          </span>
        ))}
      </td>
    </tr>
  );
};

export default AccessURLItem;
