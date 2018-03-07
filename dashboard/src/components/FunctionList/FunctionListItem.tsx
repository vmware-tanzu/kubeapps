import * as React from "react";
import { Link } from "react-router-dom";

import { IFunction } from "../../shared/types";
import Card, { CardContent } from "../Card";
import "../ChartList/ChartListItem.css";
import FunctionIcon from "./FunctionIcon";

interface IFunctionListItemProps {
  function: IFunction;
}

class FunctionListItem extends React.Component<IFunctionListItemProps> {
  public render() {
    const { function: f } = this.props;

    return (
      <Card responsive={true} className="FunctionListItem">
        <Link to={`/functions/${f.metadata.namespace}/${f.metadata.name}`}>
          <FunctionIcon runtime={f.spec.runtime} />
          <CardContent>
            <div className="ChartListItem__content">
              <h3 className="ChartListItem__content__title">{f.metadata.name}</h3>
              <div className="ChartListItem__content__info text-r">
                <p className="margin-reset type-color-light-blue">type: {f.spec.type}</p>
                <span
                  className={`ChartListItem__content__repo padding-tiny
                  padding-h-normal type-small margin-t-small`}
                >
                  {f.metadata.namespace}
                </span>
              </div>
            </div>
          </CardContent>
        </Link>
      </Card>
    );
  }
}

export default FunctionListItem;
