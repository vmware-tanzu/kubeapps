import Column from "components/js/Column";
import Row from "components/js/Row";
import * as React from "react";
import { CdsIcon } from "../../../../Clarity/clarity";
import "./SecretItemDatum.v2.css";

interface ISecretItemDatumProps {
  name: string;
  value: string;
}

function SecretItemDatum({ name, value }: ISecretItemDatumProps) {
  const [hidden, setHidden] = React.useState(true);
  const toggleDisplay = () => setHidden(!hidden);
  const decodedValue = atob(value);

  return (
    <Row>
      <Column span={1}>
        <button
          className="secret-datum-icon"
          aria-label={hidden ? "Show Secret" : "Hide Secret"}
          aria-controls={`secret-item-datum-${name}-ref`}
          aria-expanded={!hidden}
          onClick={toggleDisplay}
        >
          {hidden ? (
            <CdsIcon shape="eye" size="md" solid={true} />
          ) : (
            <CdsIcon shape="eye-hide" size="md" solid={true} />
          )}
        </button>
      </Column>
      <Column span={11}>
        <div className="secret-datum-text" id={`secret-item-datum-${name}-ref`}>
          <span>
            {name}: {hidden ? `${decodedValue.length} bytes` : <strong>{decodedValue}</strong>}
          </span>
        </div>
      </Column>
    </Row>
  );
}

export default SecretItemDatum;
