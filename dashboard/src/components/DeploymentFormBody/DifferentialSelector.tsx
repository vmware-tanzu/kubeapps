// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { DeploymentEvent } from "shared/types";
import Differential from "./Differential";

interface IDifferentialSelectorProps {
  deploymentEvent: DeploymentEvent;
  deployedValues: string;
  defaultValues: string;
  appValues: string;
}

export default function DifferentialSelector({
  deploymentEvent,
  deployedValues,
  defaultValues,
  appValues,
}: IDifferentialSelectorProps) {
  let oldValues = "";
  let emptyDiffElement = <></>;
  if (deploymentEvent === "upgrade") {
    // If there are already some deployed values (upgrade scenario)
    // We compare the values from the old release and the new one
    oldValues = deployedValues;
    emptyDiffElement = (
      <span>
        <p>
          The values you have entered to upgrade this package with are identical to the currently
          deployed ones.
        </p>
        <p>
          If you want to restore the default values provided by the package, click on the{" "}
          <i>Restore defaults</i> button below.
        </p>
      </span>
    );
  } else {
    // If it's a new deployment, we show the different from the default
    // values for the selected version
    oldValues = defaultValues || "";
    emptyDiffElement = <span>No changes detected from the package defaults.</span>;
  }
  return (
    <Differential oldValues={oldValues} newValues={appValues} emptyDiffElement={emptyDiffElement} />
  );
}
