import * as React from "react";

import placeholder from "../../placeholder.png";
import "./ChartIcon.css";

interface IChartIconProps {
  icon?: string | null;
}

class ChartIcon extends React.Component<IChartIconProps> {
  public render() {
    const { icon } = this.props;
    const iconSrc = icon ? `api/assetsvc/${icon}` : placeholder;

    return (
      <div className="ChartIcon">
        <img className="ChartIcon__img" src={iconSrc} alt="icon" />
      </div>
    );
  }
}

export default ChartIcon;
