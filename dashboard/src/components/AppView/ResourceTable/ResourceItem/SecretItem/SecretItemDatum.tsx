// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import Column from "components/js/Column";
import Row from "components/js/Row";
import React, { useEffect, useRef } from "react";
import CopyToClipboard from "react-copy-to-clipboard";
import ReactTooltip from "react-tooltip";
import "./SecretItemDatum.css";

interface ISecretItemDatumProps {
  name: string;
  value: string;
}

function SecretItemDatum({ name, value }: ISecretItemDatumProps) {
  const [hidden, setHidden] = React.useState(true);
  const [copied, setCopied] = React.useState(false);
  const toggleDisplay = () => setHidden(!hidden);
  const copyTimeout = useRef({} as NodeJS.Timeout);
  const setCopiedTrue = () => {
    setCopied(true);
  };
  useEffect(() => {
    if (copied) {
      copyTimeout.current = setTimeout(() => {
        setCopied(false);
      }, 1000);
    }
    return () => {
      if (copyTimeout.current != null) {
        clearTimeout(copyTimeout.current);
      }
    };
  }, [copied]);
  const decodedValue = Buffer.from(value, "base64").toString();

  return (
    <Row>
      <Column span={5}>
        <label htmlFor={`secret-datum-content-${name}`} className="secret-datum-text">
          {name}
        </label>
      </Column>
      <Column span={5}>
        <input
          type={hidden ? "password" : "text"}
          id={`secret-datum-content-${name}`}
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
              <CdsIcon
                shape="copy-to-clipboard"
                size="md"
                solid={true}
                aria-label={`Copy ${name} secret value to the clipboard`}
              />
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
