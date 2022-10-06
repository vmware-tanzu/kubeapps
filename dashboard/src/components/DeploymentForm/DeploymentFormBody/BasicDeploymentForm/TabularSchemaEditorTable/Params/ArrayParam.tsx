// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsInput } from "@cds/react/input";
import { CdsRange } from "@cds/react/range";
import { CdsToggle } from "@cds/react/toggle";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { useState } from "react";
import { IBasicFormParam } from "shared/types";
import { basicFormsDebounceTime } from "shared/utils";

export interface IArrayParamProps {
  id: string;
  label: string;
  type: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    param: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
}

export default function ArrayParam(props: IArrayParamProps) {
  const { id, label, type, param, handleBasicFormParamChange } = props;

  const [currentArrayItems, setCurrentArrayItems] = useState<(string | number | boolean)[]>(
    param.currentValue ? JSON.parse(param.currentValue) : [],
  );
  const [timeout, setThisTimeout] = useState({} as NodeJS.Timeout);

  const setArrayChangesInParam = () => {
    clearTimeout(timeout);
    const func = handleBasicFormParamChange(param);
    // The reference to target get lost, so we need to keep a copy
    const targetCopy = {
      currentTarget: {
        value: JSON.stringify(currentArrayItems),
        type: "change",
      },
    } as React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>;
    setThisTimeout(setTimeout(() => func(targetCopy), basicFormsDebounceTime));
  };

  const onChangeArrayItem = (index: number, value: string | number | boolean) => {
    currentArrayItems[index] = value;
    setCurrentArrayItems([...currentArrayItems]);
    setArrayChangesInParam();
  };

  const renderInput = (type: string, index: number) => {
    switch (type) {
      case "number":
      case "integer":
        return (
          <>
            <CdsInput className="self-center">
              <input
                required={param.required}
                aria-label={label}
                id={`${id}-${index}_text`}
                type="number"
                onChange={e => onChangeArrayItem(index, Number(e.currentTarget.value))}
                value={Number(currentArrayItems[index])}
                step={param.schema?.type === "integer" ? 1 : 0.1}
              />
            </CdsInput>
            <CdsRange>
              <input
                aria-label={label}
                id={`${id}-${index}_range`}
                type="range"
                onChange={e => onChangeArrayItem(index, Number(e.currentTarget.value))}
                value={Number(currentArrayItems[index])}
                step={param.schema?.type === "integer" ? 1 : 0.1}
              />
            </CdsRange>
          </>
        );
      case "boolean":
        return (
          <CdsToggle>
            <input
              required={param.required}
              aria-label={label}
              id={`${id}-${index}_toggle`}
              type="checkbox"
              onChange={e => onChangeArrayItem(index, e.currentTarget.checked)}
              checked={!!currentArrayItems[index]}
            />
          </CdsToggle>
        );

      // TODO(agamez): handle enums and objects in arrays
      default:
        return (
          <CdsInput>
            <input
              required={param.required}
              aria-label={label}
              value={currentArrayItems[index] as string}
              onChange={e => onChangeArrayItem(index, e.currentTarget.value)}
            />
          </CdsInput>
        );
    }
  };

  const onAddArrayItem = () => {
    switch (type) {
      case "number":
      case "integer":
        currentArrayItems.push(0);
        break;
      case "boolean":
        currentArrayItems.push(false);
        break;
      default:
        currentArrayItems.push("");
        break;
    }
    setCurrentArrayItems([...currentArrayItems]);
    setArrayChangesInParam();
  };

  const onDeleteArrayItem = (index: number) => {
    currentArrayItems.splice(index, 1);
    setCurrentArrayItems([...currentArrayItems]);
  };

  return (
    <>
      <CdsButton
        title={"Add a new value"}
        type="button"
        onClick={onAddArrayItem}
        action="flat"
        status="primary"
        size="sm"
      >
        <CdsIcon shape="plus" size="sm" solid={true} />
        <span>Add</span>
      </CdsButton>
      {currentArrayItems?.map((_, index) => (
        <Row key={`${id}-${index}`}>
          <Column span={9}>{renderInput(type, index)}</Column>
          <Column span={1}>
            <CdsButton
              title={"Delete"}
              type="button"
              onClick={() => onDeleteArrayItem(index)}
              action="flat"
              status="primary"
              size="sm"
            >
              <CdsIcon shape="minus" size="sm" solid={true} />
            </CdsButton>
          </Column>
        </Row>
      ))}
    </>
  );
}
