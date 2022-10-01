// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import ReactTooltip from "react-tooltip";
import { IBasicFormParam } from "shared/types";
import ArrayParam from "./Params/ArrayParam";
import BooleanParam from "./Params/BooleanParam";
import CustomFormComponentLoader from "./Params/CustomFormParam";
import SliderParam from "./Params/SliderParam";
import TextParam from "./Params/TextParam";

const MAX_LENGTH = 60;

function renderCellWithTooltip(
  value: IBasicFormParam,
  property: string,
  className = "",
  trimFromBeginning = false,
  maxLength = MAX_LENGTH,
) {
  // If the value is an object/array, we need to stringify it
  const stringValue = ["string", "number"].includes(typeof value?.[property])
    ? value?.[property] || ""
    : JSON.stringify(value?.[property]);

  if (stringValue?.length > maxLength) {
    const trimmedString = trimFromBeginning
      ? "..." + stringValue.substring(stringValue.length - maxLength, stringValue.length)
      : stringValue.substring(0, maxLength - 1) + "...";

    return (
      <span className={className}>
        <p data-tip={stringValue}>{trimmedString}</p>
        <ReactTooltip />
      </span>
    );
  } else {
    return <span className={className}>{stringValue}</span>;
  }
}

export function renderConfigKeyHeader(table: any, _saveAllChanges: any) {
  return (
    <>
      <div
        style={{
          textAlign: "left",
        }}
      >
        <>
          <CdsButton
            title={table.getIsAllRowsExpanded() ? "Collapse All" : "Expand All"}
            type="button"
            onClick={table.getToggleAllRowsExpandedHandler()}
            action="flat"
            status="primary"
            size="sm"
            className="table-button"
          >
            {table.getIsAllRowsExpanded() ? (
              <CdsIcon shape="minus" size="sm" solid={true} />
            ) : (
              <CdsIcon shape="plus" size="sm" solid={true} />
            )}
          </CdsButton>
          <span>Key</span>
        </>
      </div>
    </>
  );
}

export function renderConfigKey(value: IBasicFormParam, row: any, _saveAllChanges: any) {
  return (
    <div
      className="left-align self-center"
      style={{
        paddingLeft: `${row.depth * 0.5}rem`,
      }}
    >
      <>
        <div style={{ display: "inline-flex" }}>
          <CdsButton
            title={row.getIsExpanded() ? "Collapse" : "Expand"}
            type="button"
            onClick={row.getToggleExpandedHandler()}
            action="flat"
            status="primary"
            size="sm"
            disabled={!row.getCanExpand()}
            className="table-button"
          >
            {row.getCanExpand() ? (
              row.getIsExpanded() ? (
                <CdsIcon shape="minus" size="sm" solid={true} />
              ) : (
                <CdsIcon shape="plus" size="sm" solid={true} />
              )
            ) : (
              <></>
            )}
          </CdsButton>
          {renderCellWithTooltip(value, "key", "breakable self-center", true, MAX_LENGTH / 1.5)}
        </div>
      </>
    </div>
  );
}

export function renderConfigType(value: IBasicFormParam) {
  return renderCellWithTooltip(value, "type", "italics");
}

export function renderConfigDescription(value: IBasicFormParam) {
  return renderCellWithTooltip(value, "title", "breakable");
}

export function renderConfigDefaultValue(value: IBasicFormParam) {
  return renderCellWithTooltip(value, "defaultValue", "breakable");
}

export function renderConfigDeployedValue(value: IBasicFormParam) {
  return renderCellWithTooltip(value, "deployedValue");
}

export function renderConfigCurrentValuePro(
  param: IBasicFormParam,
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void,
) {
  // early return if the value is marked as a custom form component
  if (param.isCustomComponent) {
    // TODO(agamez): consider using a modal window to display the full value
    return (
      <div id={param.key}>
        <CustomFormComponentLoader
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />
      </div>
    );
  }
  // if the param has properties, each of them will be rendered as a row
  if (param.hasProperties) {
    return <></>;
  }

  // if it isn't a custom component or an with more properties, render an input
  switch (param.type) {
    case "string":
      return (
        <TextParam
          id={param.key}
          label={param.title || param.path}
          param={param}
          inputType={
            [param.title, param.key].some(s => s.match(/password/i)) ? "password" : "string"
          }
          handleBasicFormParamChange={handleBasicFormParamChange}
        />
      );

    case "boolean":
      return (
        <BooleanParam
          id={param.key}
          label={param.title || param.path}
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />
      );

    case "integer":
    case "number":
      return (
        <SliderParam
          id={param.key}
          label={param.title || param.path}
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
          step={param.type === "integer" ? 1 : 0.1}
          unit={""}
        />
      );
    case "array":
      if (param?.schema?.items?.type !== "object") {
        return (
          <ArrayParam
            id={param.key}
            label={param.title || param.path}
            param={param}
            handleBasicFormParamChange={handleBasicFormParamChange}
            type={param?.schema?.items?.type ?? "string"}
          />
        );
      } else {
        // TODO(agamez): render the object properties
        return (
          <TextParam
            id={param.key}
            label={param.title || param.path}
            param={param}
            inputType={"textarea"}
            handleBasicFormParamChange={handleBasicFormParamChange}
          />
        );
      }
    default:
      return (
        <TextParam
          id={param.key}
          label={param.title || param.path}
          param={param}
          inputType={"textarea"}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />
      );
  }
}
