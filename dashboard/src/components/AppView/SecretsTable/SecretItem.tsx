import * as _ from "lodash";
import * as React from "react";
import { AlertTriangle } from "react-feather";

import LoadingWrapper, { LoaderType } from "../../../components/LoadingWrapper";
import { IKubeItem, ISecret } from "../../../shared/types";
import "./SecretContent.css";
import SecretItemDatum from "./SecretItemDatum";

interface ISecretItemProps {
  name: string;
  secret?: IKubeItem<ISecret>;
  getSecret: () => void;
}

class SecretItem extends React.Component<ISecretItemProps> {
  public componentDidMount() {
    this.props.getSecret();
  }

  public render() {
    const { name, secret } = this.props;
    return (
      <tr className="flex">
        <td className="col-3">{name}</td>
        {this.renderSecretInfo(secret)}
      </tr>
    );
  }

  private renderSecretInfo(secret?: IKubeItem<ISecret>) {
    if (secret === undefined || secret.isFetching) {
      return (
        <td className="col-9">
          <LoadingWrapper type={LoaderType.Placeholder} />
        </td>
      );
    }
    if (secret.error) {
      return (
        <td className="col-9">
          <span className="flex">
            <AlertTriangle />
            <span className="flex margin-l-normal">Error: {secret.error.message}</span>
          </span>
        </td>
      );
    }
    if (secret.item) {
      const item = secret.item;
      return (
        <React.Fragment>
          <td className="col-2">{item.type}</td>
          <td className="col-7 padding-small">
            {item.data ? (
              Object.keys(item.data).map(k => (
                <SecretItemDatum key={`${item.metadata.name}/${k}`} name={k} value={item.data[k]} />
              ))
            ) : (
              <span>This Secret is empty</span>
            )}
          </td>
        </React.Fragment>
      );
    }
    return null;
  }
}

export default SecretItem;
