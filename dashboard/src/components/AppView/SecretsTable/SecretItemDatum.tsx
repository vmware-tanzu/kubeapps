import * as React from "react";
import { Eye, EyeOff } from "react-feather";

interface ISecretItemDatumProps {
  name: string;
  value: string;
}

interface ISecretItemDatumState {
  hidden: boolean;
}

class SecretItemDatum extends React.PureComponent<ISecretItemDatumProps, ISecretItemDatumState> {
  // Secret datum is hidden by default
  public state: ISecretItemDatumState = {
    hidden: true,
  };

  public render() {
    const { name, value } = this.props;
    const { hidden } = this.state;
    const decodedValue = atob(value);
    return (
      <span className="flex">
        <a onClick={this.toggleDisplay}>{hidden ? <Eye /> : <EyeOff />}</a>
        <span className="flex margin-l-normal">
          <span>{name}:</span>
          {hidden ? (
            <span className="margin-l-small">{decodedValue.length} bytes</span>
          ) : (
            <pre className="SecretContainer">
              <code className="SecretContent">{decodedValue}</code>
            </pre>
          )}
        </span>
      </span>
    );
  }

  private toggleDisplay = () => {
    this.setState({
      hidden: !this.state.hidden,
    });
  };
}

export default SecretItemDatum;
