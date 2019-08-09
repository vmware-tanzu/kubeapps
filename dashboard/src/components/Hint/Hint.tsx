import * as React from "react";
import { HelpCircle } from "react-feather";
import * as ReactTooltip from "react-tooltip";

import "./Hint.css";

interface IHintProps {
  reactTooltipOpts?: any;
  id?: string;
}

export class Hint extends React.Component<IHintProps> {
  public render() {
    // Generate a random ID if not given
    const id =
      this.props.id ||
      Math.random()
        .toString(36)
        .substring(7);
    return (
      <React.Fragment>
        <a data-tip={true} data-for={id}>
          <HelpCircle className="icon" color="white" fill="#5F6369" />
        </a>
        <ReactTooltip
          id={id}
          className="extraClass"
          effect="solid"
          {...this.props.reactTooltipOpts}
        >
          {this.props.children}
        </ReactTooltip>
      </React.Fragment>
    );
  }
}

export default Hint;
