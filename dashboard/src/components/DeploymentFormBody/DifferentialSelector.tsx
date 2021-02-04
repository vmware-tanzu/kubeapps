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
  let title = "";
  let emptyDiffText = "";
  if (deploymentEvent === "upgrade") {
    // If there are already some deployed values (upgrade scenario)
    // We compare the values from the old release and the new one
    oldValues = deployedValues;
    title = "Difference from deployed version";
    emptyDiffText = "The values for the new release are identical to the deployed version.";
  } else {
    // If it's a new deployment, we show the different from the default
    // values for the selected version
    oldValues = defaultValues || "";
    title = "Difference from chart defaults";
    emptyDiffText = "No changes detected from chart defaults.";
  }
  return (
    <Differential
      title={title}
      oldValues={oldValues}
      newValues={appValues}
      emptyDiffText={emptyDiffText}
    />
  );
}
