import * as React from "react";

import "./CardIcon.css";

export interface ICardIconProps {
  icon?: string;
}

class CardIcon extends React.PureComponent<ICardIconProps> {
  public render() {
    const { icon } = this.props;

    return icon && icon !== "" ? (
      <div className="Card__icon bg-light text-c">
        <img src={icon} />
      </div>
    ) : (
      ""
    );
  }
}

export default CardIcon;
