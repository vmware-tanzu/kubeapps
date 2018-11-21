import * as React from "react";
import { Eye, EyeOff } from "react-feather";

import { ISecret } from "../../shared/types";
import "./SecretContent.css";

interface ISecretItemProps {
  secret: ISecret;
}

interface ISecretItemState {
  display: { [s: string]: boolean };
}

class SecretItem extends React.Component<ISecretItemProps, ISecretItemState> {
  public state: ISecretItemState = {
    display: {},
  };

  public componentDidMount() {
    const display = {};
    Object.keys(this.props.secret.data).forEach(k => (display[k] = false));
    this.setState(display);
  }

  public render() {
    const { secret } = this.props;
    let data: JSX.Element[] = [];
    Object.keys(secret.data).forEach(k => {
      data = data.concat(this.renderSecret(k));
    });
    return (
      <tr className="flex">
        <td className="col-2">{secret.metadata.name}</td>
        <td className="col-2">{secret.type}</td>
        <td className="col-7 padding-small">{data}</td>
      </tr>
    );
  }

  private renderSecret = (name: string) => {
    const toggle = () => this.toggleDisplay(name);
    const text = atob(this.props.secret.data[name]);
    return (
      <span className="flex">
        <a onClick={toggle}>{this.state.display[name] ? <EyeOff /> : <Eye />}</a>
        <span className="flex margin-l-normal">
          <span>{name}:</span>
          {this.state.display[name] ? (
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
    const { display } = this.state;
    display[name] = !display[name];
    this.setState({ display });
  };
}

export default SecretItem;
