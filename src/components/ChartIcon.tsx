import * as React from 'react';

const placeholder = require('../placeholder.png');

interface Props {
  icon: string;
}

class ChartIcon extends React.Component<Props> {
  render() {
    const { icon } = this.props;
    let iconSrc;

    if (icon.length > 0) {
      iconSrc = `/api/chartsvc/${icon}`;
    } else {
      iconSrc = placeholder;
    }

    return (
      <div className="ChartListItem__icon">
        <img className="ChartListItem__icon" src={iconSrc} />
      </div>
    );
  }
}

export default ChartIcon;
