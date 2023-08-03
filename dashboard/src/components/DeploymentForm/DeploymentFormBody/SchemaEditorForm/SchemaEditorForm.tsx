// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsControlMessage } from "@cds/react/forms";
import { CdsIcon } from "@cds/react/icon";
import { CdsRadio, CdsRadioGroup } from "@cds/react/radio";
import Column from "components/Column";
import Row from "components/Row";
import { isEmpty } from "lodash";
import { useEffect, useState } from "react";
import { MonacoDiffEditor, monaco } from "react-monaco-editor";
import { useSelector } from "react-redux";
import { schemaToObject, validateSchema } from "shared/schema";
import { IAjvValidateResult, IStoreState } from "shared/types";

export interface ISchemaEditorForm {
  schemaFromTheParentContainer?: string;
  handleValuesChange: (value: string) => void;
  children?: JSX.Element;
  schemaFromTheAvailablePackage: string;
}

export default function SchemaEditorForm(props: ISchemaEditorForm) {
  const {
    config: { theme },
  } = useSelector((state: IStoreState) => state);
  const { handleValuesChange, schemaFromTheParentContainer, schemaFromTheAvailablePackage } = props;

  const [useDiffEditor, setUseDiffEditor] = useState(false);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [diffValues, setDiffValues] = useState(schemaFromTheAvailablePackage);
  const [currentSchema, setCurrentSchema] = useState(schemaFromTheParentContainer || "{}");
  const [validated, setValidated] = useState<IAjvValidateResult>();

  const diffEditorOptions = {
    renderSideBySide: false,
    automaticLayout: true,
  };

  const validateSchemaInput = (schema?: string) => {
    const parsedSchema = schemaToObject(schema);
    const schemaValidation = validateSchema(parsedSchema);
    if (schema !== "{}" && isEmpty(parsedSchema)) {
      schemaValidation.valid = false;
      schemaValidation.errors = [
        { message: "Invalid JSON", instancePath: "", keyword: "", params: {}, schemaPath: "" },
      ];
    }
    return schemaValidation;
  };

  const onChange = (value: string | undefined, _ev: any) => {
    setCurrentSchema(value || "{}");
    setHasUnsavedChanges(true);
    setValidated(validateSchemaInput(value));
  };

  const handleUpdateSchema = (currentSchema?: string) => {
    const schemaValidation = validateSchemaInput(currentSchema);
    setValidated(schemaValidation);
    if (currentSchema && schemaValidation.valid) {
      handleValuesChange(currentSchema);
      setHasUnsavedChanges(false);
    }
  };

  useEffect(() => {
    if (!useDiffEditor) {
      setDiffValues(currentSchema);
    } else {
      setDiffValues(schemaFromTheParentContainer || "");
    }
  }, [currentSchema, schemaFromTheParentContainer, useDiffEditor]);

  const editorDidMount = (editor: monaco.editor.IStandaloneDiffEditor, m: typeof monaco) => {
    // Add "update schema" action
    editor.addAction({
      id: "updateSchema",
      label: "Update schema",
      keybindings: [m.KeyMod.CtrlCmd | m.KeyCode.KeyS],
      contextMenuGroupId: "9_cutcopypaste",
      run: () => {
        handleUpdateSchema(editor.getModel()?.modified.getValue());
      },
    });
  };

  return (
    <div className="deployment-form-tabs-data">
      <div className="deployment-form-tabs-data-buttons">
        <Row>
          <Column span={3}>
            <CdsRadioGroup layout="vertical">
              {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
              <label>Enable diff editor:</label>
              <CdsRadio>
                <label htmlFor="diff-compare-enable-true">Yes</label>
                <input
                  id="diff-compare-enable-true"
                  type="radio"
                  name="true"
                  checked={useDiffEditor}
                  onChange={e => {
                    setUseDiffEditor(e.target.checked);
                  }}
                />
              </CdsRadio>
              <CdsRadio>
                <label htmlFor="diff-compare-enable-false">No</label>
                <input
                  id="diff-compare-enable-false"
                  type="radio"
                  name="deployed"
                  checked={!useDiffEditor}
                  onChange={e => {
                    setUseDiffEditor(!e.target.checked);
                  }}
                />
              </CdsRadio>
            </CdsRadioGroup>
          </Column>
          <>
            {hasUnsavedChanges && (
              <Column span={3}>
                <CdsButton
                  type="button"
                  status="primary"
                  action="outline"
                  onClick={() => handleUpdateSchema(currentSchema)}
                  disabled={!validated?.valid}
                >
                  <CdsIcon shape="upload" /> Update schema
                </CdsButton>
                <>
                  {!validated?.valid && !isEmpty(validated?.errors) && (
                    <CdsControlMessage status="error">
                      {validated?.errors?.map((e: any) => e?.message).join(", ")}
                    </CdsControlMessage>
                  )}
                </>
              </Column>
            )}
          </>
        </Row>
      </div>
      <br />
      <div className="deployment-form-tabs-data schema-editor">
        <MonacoDiffEditor
          value={currentSchema}
          original={diffValues}
          height="90vh"
          language="json"
          theme={theme === "dark" ? "vs-dark" : "light"}
          options={diffEditorOptions}
          onChange={onChange}
          editorDidMount={editorDidMount}
        />
      </div>
    </div>
  );
}
