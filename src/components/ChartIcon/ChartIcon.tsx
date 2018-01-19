import * as React from "react";
import placeholder from "../../placeholder.png";

interface IChartIconProps {
  icon?: string;
}

class ChartIcon extends React.Component<IChartIconProps> {
  public render() {
    const { icon } = this.props;
    const iconSrc = icon ? `/api/chartsvc/${icon}` : placeholder;

    return (
      <div className="ChartListItem__icon">
        <img className="ChartListItem__icon" src={iconSrc} />
      </div>
    );
  }
}

export default ChartIcon;
