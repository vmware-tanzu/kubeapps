// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";

interface IOperatorCapabilitiesProps {
  level: string;
}

export const BASIC_INSTALL = "Basic Install";
export const SEAMLESS_UPGRADES = "Seamless Upgrades";
export const FULL_LIFECYCLE = "Full Lifecycle";
export const DEEP_INSIGHTS = "Deep Insights";
export const AUTO_PILOT = "Auto Pilot";

const levels = {
  [BASIC_INSTALL]: 1,
  [SEAMLESS_UPGRADES]: 2,
  [FULL_LIFECYCLE]: 3,
  [DEEP_INSIGHTS]: 4,
  [AUTO_PILOT]: 5,
};

export default function CapabiliyLevel({ level }: IOperatorCapabilitiesProps) {
  const levelInt = levels[level];
  return (
    <ul className="capabilities-list">
      {Object.keys(levels).map(key => {
        return (
          <li key={key}>
            {levels[key] <= levelInt ? (
              <span className="color-icon-info">
                <CdsIcon shape="success-standard" solid={true} />
              </span>
            ) : (
              <CdsIcon shape="times-circle" solid={true} />
            )}
            <span style={{ color: levels[key] <= levelInt ? "inherit" : "#C7C9C8" }}>{key}</span>
          </li>
        );
      })}
    </ul>
  );
}
