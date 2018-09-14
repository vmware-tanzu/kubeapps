import * as React from "react";

import "./Card.css";

export interface ICardProps {
  className?: string;
  children?: React.ReactChildren | React.ReactNode | string;
  responsive?: boolean;
  responsiveColumns?: number;
}

class Card extends React.Component<ICardProps> {
  public cssClass() {
    let cssClass = `Card ${this.props.className || ""}`;

    if (this.props.responsive && this.props.responsiveColumns) {
      cssClass = `${cssClass} Card-responsive-${this.props.responsiveColumns}`;
    } else if (this.props.responsive) {
      cssClass = `${cssClass} Card-responsive`;
    }
    return cssClass;
  }

  public render() {
    return (
      <article className={this.cssClass()}>
        <div className="Card__inner bg-white elevation-1">{this.props.children}</div>
      </article>
    );
  }
}

export default Card;
