import * as React from "react";

import { IFunction } from "shared/types";
import Card, { CardContent, CardFooter, CardGrid } from "../Card";
import FunctionIcon from "../FunctionIcon";

interface IFunctionInfoProps {
  function: IFunction;
}

class FunctionInfo extends React.Component<IFunctionInfoProps> {
  public render() {
    const { function: f } = this.props;
    const name = f.metadata.name;
    return (
      <CardGrid className="FunctionInfo">
        <Card>
          <FunctionIcon runtime={f.spec.runtime} />
          <CardContent className="padding-v-reset">
            <h5>{name}</h5>
          </CardContent>
          <CardFooter>
            <ul className="remove-style margin-reset padding-reset type-small">
              <li>handler: {f.spec.handler}</li>
              <li>runtime: {f.spec.runtime}</li>
            </ul>
          </CardFooter>
        </Card>
      </CardGrid>
    );
  }
}

export default FunctionInfo;
