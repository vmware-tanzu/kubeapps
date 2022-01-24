// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import { useEffect, useState } from "react";
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
      // We compare the values from the previously deployed release and the new one
      setOldValues(deployedValues);
    } else {
      // If it's a new deployment, we show the difference from the default
      // values for the selected version
      setOldValues(defaultValues || "");
    }
  }, [deployedValues, defaultValues, deploymentEvent]);
  useEffect(() => {
    if (oldValues !== "") {
      if (oldValues !== appValues) {
        setNewChanges(true);
      } else {
        setNewChanges(false);
      }
    }
  }, [oldValues, appValues]);
  return (
    <div role="presentation" onClick={setNewChangesFalse} className="notification-icon">
      Changes
      <CdsIcon hidden={!newChanges} shape="circle" solid={true} />
    </div>
  );
}
