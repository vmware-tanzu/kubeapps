import * as React from "react";
import { CheckCircle, XCircle } from "react-feather";

import "./CapabilityLevel.css";

interface IOperatorCapabilitiesProps {
  level: string;
}

const levels = {
  "Basic Install": 1,
  "Seamless Upgrades": 2,
  "Full Lifecycle": 3,
  "Deep Insights": 4,
  "Auto Pilot": 5,
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
