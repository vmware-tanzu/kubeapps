import * as React from "react";

import nodeIcon from "../../img/node.png";
import pythonIcon from "../../img/python.png";
import rubyIcon from "../../img/ruby.png";
import placeholder from "../../placeholder.png";
import { CardIcon } from "../Card";

interface IFunctionIconProps {
  runtime: string;
}

class FunctionIcon extends React.Component<IFunctionIconProps> {
  public render() {
    const { runtime } = this.props;
    let src = placeholder;
    if (runtime.match(/node/)) {
      src = nodeIcon;
    } else if (runtime.match(/ruby/)) {
      src = rubyIcon;
    } else if (runtime.match(/python/)) {
      src = pythonIcon;
    }
    return <CardIcon icon={src} />;
  }
}

export default FunctionIcon;
