// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsRadio, CdsRadioGroup } from "@cds/react/radio";
import Column from "components/Column";
import Row from "components/Row";
import monaco from "monaco-editor/esm/vs/editor/editor.api"; // for types only
import { useEffect, useState } from "react";
import { MonacoDiffEditor } from "react-monaco-editor";
import { useSelector } from "react-redux";
import { IStoreState } from "shared/types";

export interface IAdvancedDeploymentForm {
  valuesFromTheParentContainer?: string;
  handleValuesChange: (value: string) => void;
  children?: JSX.Element;
  valuesFromTheDeployedPackage: string;
  valuesFromTheAvailablePackage: string;
  deploymentEvent: string;
}

export default function AdvancedDeploymentForm(props: IAdvancedDeploymentForm) {
  const {
    config: { theme },
  } = useSelector((state: IStoreState) => state);
  const {
    handleValuesChange,
    valuesFromTheParentContainer,
    valuesFromTheDeployedPackage,
    valuesFromTheAvailablePackage,
    deploymentEvent,
  } = props;

  const [usePackageDefaults, setUsePackageDefaults] = useState(
    deploymentEvent === "upgrade" ? false : true,
  );
  const [useDiffEditor, setUseDiffEditor] = useState(true);
  const [diffValues, setDiffValues] = useState(valuesFromTheAvailablePackage);

  const diffEditorOptions = {
    renderSideBySide: false,
    automaticLayout: true,
  };

  const onChange = (value: string | undefined, _ev: any) => {
    // debouncing is not required as the diff calculation happens in a webworker
    handleValuesChange(value || "");
  };

  useEffect(() => {
    if (!useDiffEditor) {
      setDiffValues(valuesFromTheParentContainer || "");
    } else if (!usePackageDefaults) {
      setDiffValues(valuesFromTheDeployedPackage);
    } else {
      setDiffValues(valuesFromTheAvailablePackage);
    }
  }, [
    usePackageDefaults,
    useDiffEditor,
    valuesFromTheAvailablePackage,
    valuesFromTheDeployedPackage,
    valuesFromTheParentContainer,
  ]);

  const editorDidMount = (editor: monaco.editor.IStandaloneDiffEditor, m: typeof monaco) => {
    // Add "go to the next change" action
    editor.addAction({
      id: "goToNextChange",
      label: "Go to the next change",
      keybindings: [m.KeyMod.Alt | m.KeyCode.KeyG],
      contextMenuGroupId: "9_cutcopypaste",
      run: () => {
        const lineChanges = editor?.getLineChanges() as monaco.editor.ILineChange[];
        lineChanges.some(lineChange => {
          const currentPosition = editor?.getPosition() as monaco.Position;
          if (currentPosition.lineNumber < lineChange.modifiedEndLineNumber) {
            // Set the cursor to the next change
            editor?.setPosition({
              lineNumber: lineChange.modifiedEndLineNumber,
              column: 1,
            });
            // Scroll to the next change
            editor?.revealPositionInCenter({
              lineNumber: lineChange.modifiedEndLineNumber,
              column: 1,
            });
            // Return true to stop the loop
            return true;
          }
          return false;
        });
      },
    });
    // Add "go to the previous change" action
    editor.addAction({
      id: "goToPreviousChange",
      label: "Go to the previous change",
      keybindings: [m.KeyMod.Alt | m.KeyCode.KeyF],
      contextMenuGroupId: "9_cutcopypaste",
      run: () => {
        const lineChanges = editor?.getLineChanges() as monaco.editor.ILineChange[];
        lineChanges.some(lineChange => {
          const currentPosition = editor?.getPosition() as monaco.Position;
          if (currentPosition.lineNumber > lineChange.modifiedEndLineNumber) {
            // Set the cursor to the next change
            editor?.setPosition({
              lineNumber: lineChange.modifiedEndLineNumber,
              column: 1,
            });
            // Scroll to the next change
            editor?.revealPositionInCenter({
              lineNumber: lineChange.modifiedEndLineNumber,
              column: 1,
            });
            // Return true to stop the loop
            return true;
          }
          return false;
        });
      },
    });

    // Add the "toggle deployed/package default values" action
    if (deploymentEvent === "upgrade") {
      editor.addAction({
        id: "useDefaultsFalse",
        label: "Use default values",
        keybindings: [m.KeyMod.Alt | m.KeyCode.KeyD],
        contextMenuGroupId: "9_cutcopypaste",
        run: () => {
          setUsePackageDefaults(false);
        },
      });
      editor.addAction({
        id: "useDefaultsTrue",
        label: "Use package values",
        keybindings: [m.KeyMod.Alt | m.KeyCode.KeyV],
        contextMenuGroupId: "9_cutcopypaste",
        run: () => {
          setUsePackageDefaults(true);
        },
      });
    }
  };

  return (
    <div className="deployment-form-tabs-data">
      <div className="deployment-form-tabs-data-buttons">
        <Row>
          <Column span={3}>
            <div className="deployment-form-tabs-buttons">
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
            </div>
          </Column>
          {deploymentEvent === "upgrade" ? (
            <>
              <Column span={3}>
                <CdsRadioGroup layout="vertical">
                  {/* eslint-disable-next-line jsx-a11y/label-has-associated-control */}
                  <label>Values to compare against:</label>
                  <CdsRadio>
                    <label htmlFor="diff-compare-values-package">Package values</label>
                    <input
                      id="diff-compare-values-package"
                      type="radio"
                      name="package"
                      checked={usePackageDefaults}
                      onChange={e => {
                        setUsePackageDefaults(e.target.checked);
                      }}
                    />
                  </CdsRadio>
                  <CdsRadio>
                    <label htmlFor="diff-compare-values-deployed">Deployed values</label>
                    <input
                      id="diff-compare-values-deployed"
                      type="radio"
                      name="deployed"
                      checked={!usePackageDefaults}
                      onChange={e => {
                        setUsePackageDefaults(!e.target.checked);
                      }}
                    />
                  </CdsRadio>
                </CdsRadioGroup>
              </Column>
            </>
          ) : (
            <></>
          )}
        </Row>
      </div>
      <br />
      <div className="deployment-form-tabs-data values-editor">
        <MonacoDiffEditor
          value={valuesFromTheParentContainer}
          original={diffValues}
          height="90vh"
          language="yaml"
          theme={theme === "dark" ? "vs-dark" : "light"}
          options={diffEditorOptions}
          onChange={onChange}
          editorDidMount={editorDidMount}
        />
      </div>
    </div>
  );
}
