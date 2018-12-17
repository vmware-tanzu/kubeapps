import * as _ from "lodash";
import * as React from "react";
import { Eye, EyeOff } from "react-feather";

import { ISecret } from "../../../shared/types";
import "./SecretContent.css";

interface ISecretItemProps {
  secret: ISecret;
}

interface ISecretItemState {
  showSecret: { [s: string]: boolean };
}

class SecretItem extends React.Component<ISecretItemProps, ISecretItemState> {
  public constructor(props: ISecretItemProps) {
    super(props);
    const showSecret = {};
    if (this.props.secret.data) {
      Object.keys(this.props.secret.data).forEach(k => (showSecret[k] = false));
    }
    this.state = { showSecret };
  }

  public render() {
    const { secret } = this.props;
    const secretEntries: JSX.Element[] = [];
    if (!_.isEmpty(this.props.secret.data)) {
      Object.keys(secret.data).forEach(k => {
        secretEntries.push(this.renderSecretEntry(k));
      });
    } else {
      secretEntries.push(<span key="empty">The secret is empty</span>);
    }
    return (
      <tr className="flex">
        <td className="col-2">{secret.metadata.name}</td>
        <td className="col-2">{secret.type}</td>
        <td className="col-7 padding-small">{secretEntries}</td>
      </tr>
    );
  }

  private renderSecretEntry = (name: string) => {
    const toggle = () => this.toggleDisplay(name);
    const text = atob(this.props.secret.data[name]);
    return (
      <span key={name} className="flex">
        <a onClick={toggle}>{this.state.showSecret[name] ? <EyeOff /> : <Eye />}</a>
        <span className="flex margin-l-normal">
          <span>{name}:</span>
          {this.state.showSecret[name] ? (
            <pre className="SecretContainer">
              <code className="SecretContent">{text}</code>
            </pre>
          ) : (
            <span className="margin-l-small">{text.length} bytes</span>
          )}
        </span>
      </span>
    );
  };

  private toggleDisplay = (name: string) => {
    const { showSecret } = this.state;
    this.setState({ showSecret: { ...showSecret, [name]: !showSecret[name] } });
  };
}

export default SecretItem;
