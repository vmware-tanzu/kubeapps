// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import ReactDiffViewer, { DiffMethod } from "react-diff-viewer";
import { useSelector } from "react-redux";
import { SupportedThemes } from "shared/Config";
import { IStoreState } from "shared/types";
import "./Differential.css";

export interface IDifferentialProps {
  oldValues: string;
  newValues: string;
  emptyDiffElement: JSX.Element;
}

function Differential(props: IDifferentialProps) {
  const { oldValues, newValues, emptyDiffElement } = props;
  const {
    config: { theme },
  } = useSelector((state: IStoreState) => state);

  // Modify colors to match the Advanced Tab theme
  const newStyles = {
    variables: {
      dark: {
        // gutter
        gutterColor: "#d0edf7",
        gutterBackground: "#01313f",
        gutterBackgroundDark: "#01313f",
        addedGutterColor: "#d0edf7",
        removedGutterColor: "#d0edf7",
        // background
        diffViewerBackground: "#002B36",
        emptyLineBackground: "#002B36",
        // fold text
        codeFoldContentColor: "white",
      },
    },
  };

  return (
    <div className="diff deployment-form-tabs-data">
      {oldValues === newValues ? (
        emptyDiffElement
      ) : (
        <ReactDiffViewer
          oldValue={oldValues}
          newValue={newValues}
          splitView={false}
          useDarkTheme={theme === SupportedThemes.dark}
          compareMethod={DiffMethod.WORDS}
          styles={newStyles}
        />
      )}
    </div>
  );
}

export default Differential;
