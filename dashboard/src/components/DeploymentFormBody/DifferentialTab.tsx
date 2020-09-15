import { CdsIcon } from "@clr/react/icon";
import React, { useEffect, useState } from "react";
import { DeploymentEvent } from "shared/types";

interface IDifferentialSelectorProps {
  deploymentEvent: DeploymentEvent;
  deployedValues: string;
  defaultValues: string;
  appValues: string;
}

export default function DifferentialTab({
  deploymentEvent,
  deployedValues,
  defaultValues,
  appValues,
}: IDifferentialSelectorProps) {
  const [newChanges, setNewChanges] = useState(false);
  const [oldValues, setOldValues] = useState("");
  const setNewChangesFalse = () => setNewChanges(false);

  useEffect(() => {
    if (deploymentEvent === "upgrade") {
      // If there are already some deployed values (upgrade scenario)
      // We compare the values from the old release and the new one
      setOldValues(deployedValues);
    } else {
      // If it's a new deployment, we show the different from the default
      // values for the selected version
      setOldValues(defaultValues || "");
    }
  }, [deployedValues, defaultValues, deploymentEvent]);
  useEffect(() => {
    if (oldValues !== "" && oldValues !== appValues) {
      setNewChanges(true);
    }
  }, [oldValues, appValues]);
  return (
    <div onClick={setNewChangesFalse} className="notification-icon">
      Changes
      <CdsIcon hidden={!newChanges} shape="circle" solid={true} />
    </div>
  );
}
