import { ClarityIcons, eyeHideIcon, eyeIcon } from "@clr/core/icon-shapes";
import Column from "components/js/Column";
import Row from "components/js/Row";
import * as React from "react";
import { CdsIcon } from "../../../../Clarity/clarity";
import "./SecretItemDatum.v2.css";
ClarityIcons.addIcons(eyeIcon, eyeHideIcon);

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
      <Row>
        <Column span={1}>
          <button className="secret-datum-icon" onClick={this.toggleDisplay}>
            {hidden ? (
              <CdsIcon shape="eye" size="md" solid={true} />
            ) : (
              <CdsIcon shape="eye-hide" size="md" solid={true} />
            )}
          </button>
        </Column>
        <Column span={11}>
          <div className="secret-datum-text">
            {name}: {hidden ? `${decodedValue.length} bytes` : `${decodedValue}`}
          </div>
        </Column>
      </Row>
    );
  }

  private toggleDisplay = () => {
    this.setState({
      hidden: !this.state.hidden,
    });
  };
}

export default SecretItemDatum;
