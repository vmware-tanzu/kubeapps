import { ClarityIcons, copyToClipboardIcon, eyeHideIcon, eyeIcon } from "@clr/core/icon-shapes";
import Column from "components/js/Column";
import Row from "components/js/Row";
import * as React from "react";
import CopyToClipboard from "react-copy-to-clipboard";
import { CdsIcon } from "../../../../Clarity/clarity";

import ReactTooltip from "react-tooltip";
import "./SecretItemDatum.v2.css";
ClarityIcons.addIcons(eyeIcon, eyeHideIcon, copyToClipboardIcon);

interface ISecretItemDatumProps {
  name: string;
  value: string;
}

function SecretItemDatum({ name, value }: ISecretItemDatumProps) {
  const [hidden, setHidden] = React.useState(true);
  const [copied, setCopied] = React.useState(false);
  const toggleDisplay = () => setHidden(!hidden);
  const setCopiedTrue = () => {
    setCopied(true);
    setTimeout(() => {
      setCopied(false);
    }, 1000);
  };
  const decodedValue = atob(value);

  return (
    <Row>
      <Column span={5}>
        <div className="secret-datum-text">{name}</div>
      </Column>
      <Column span={5}>
        <input
          type={hidden ? "password" : "text"}
          className="clr-input secret-datum-content"
          value={decodedValue}
          readOnly={true}
        />
      </Column>
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
      <Column span={1}>
        <button className="secret-datum-icon" aria-expanded={!hidden} onClick={setCopiedTrue}>
          <div data-tip={true} data-for="app-status">
            <CopyToClipboard text={decodedValue}>
              <CdsIcon shape="copy-to-clipboard" size="md" solid={true} />
            </CopyToClipboard>
          </div>
        </button>
        <div style={{ opacity: copied ? "1" : "0" }}>
          <ReactTooltip id="app-status" effect="solid">
            Copied
          </ReactTooltip>
        </div>
      </Column>
    </Row>
  );
}

export default SecretItemDatum;
