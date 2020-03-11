import * as React from "react";
import { CheckCircle, XCircle } from "react-feather";

import "./CapabilityLevel.css";

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

class CapabiliyLevel extends React.Component<IOperatorCapabilitiesProps> {
  public render() {
    const { level } = this.props;
    const levelInt = levels[level];
    return Object.keys(levels).map(key => {
      return (
        <li key={key}>
          {levels[key] <= levelInt ? (
            <CheckCircle className="capabilityLevelIcon" stroke="#1598CB" />
          ) : (
            <XCircle className="capabilityLevelIcon" stroke="#C7C9C8" />
          )}
          <span style={{ color: levels[key] <= levelInt ? "inherit" : "#C7C9C8" }}>{key}</span>
        </li>
      );
    });
  }
}

export default CapabiliyLevel;
