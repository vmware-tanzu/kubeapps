import * as React from "react";

import "./CardGrid.css";

export interface ICardGridProps {
  className?: string;
  children?: React.ReactChildren | React.ReactNode | string;
}

class CardGrid extends React.Component<ICardGridProps> {
  public render() {
    return (
      <div className={`CardGrid padding-v-big ${this.props.className || ""}`}>
        {this.props.children}
      </div>
    );
  }
}

export default CardGrid;
