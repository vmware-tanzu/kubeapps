import * as React from "react";

export interface ICardFooterProps {
  className?: string;
  children?: React.ReactChildren | React.ReactNode | string;
}

class CardFooter extends React.Component<ICardFooterProps> {
  public render() {
    return (
      <div className={`Card__footer padding-h-normal ${this.props.className || ""}`}>
        <hr className="separator-small" />
        <div className="padding-b-normal">{this.props.children}</div>
      </div>
    );
  }
}

export default CardFooter;
