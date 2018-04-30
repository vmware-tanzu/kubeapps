import * as React from "react";
import { Link } from "react-router-dom";

import { IFunction } from "../../shared/types";
import Card, { CardContent } from "../Card";
import "../ChartList/ChartListItem.css";
import FunctionIcon from "../FunctionIcon";

interface IFunctionListItemProps {
  function: IFunction;
}

class FunctionListItem extends React.Component<IFunctionListItemProps> {
  public render() {
    const { function: f } = this.props;

    return (
      <Card responsive={true} className="FunctionListItem">
        <Link
          to={`/functions/ns/${f.metadata.namespace}/${f.metadata.name}`}
          title={f.metadata.name}
        >
          <FunctionIcon runtime={f.spec.runtime} />
          <CardContent>
            <div className="ChartListItem__content">
              <h3 className="ChartListItem__content__title type-big">{f.metadata.name}</h3>
              <div className="ChartListItem__content__info">
                <p className="ChartListItem__content__info_version margin-reset type-small padding-t-tiny type-color-light-blue">
                  {" "}
                </p>
                <span
                  className={`ChartListItem__content__info_repo ${
                    f.metadata.namespace
                  } type-small padding-t-tiny padding-h-normal`}
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
