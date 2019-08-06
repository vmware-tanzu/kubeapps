import * as React from "react";
import { HelpCircle } from "react-feather";
import * as ReactTooltip from "react-tooltip";

import "./Hint.css";

interface IHintProps {
  reactTooltipOpts?: any;
}

export class Hint extends React.Component<IHintProps> {
  public render() {
    return (
      <React.Fragment>
        <a data-tip={true} data-for="syncJobHelp">
          <HelpCircle className="icon" color="white" fill="#5F6369" />
        </a>
        <ReactTooltip
          id="syncJobHelp"
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
