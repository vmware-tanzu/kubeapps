import { RouterAction } from "connected-react-router";
import * as React from "react";
import { Link } from "react-router-dom";
import { app } from "../../shared/url";

interface IOperatorHeaderProps {
  id: string;
  icon?: string;
  hideButton?: boolean;
  disableButton?: boolean;
  description: string;
  cluster: string;
  namespace: string;
  version: string;
  provider: string;
  namespaced: boolean;
  push: (location: string) => RouterAction;
}

class OperatorHeader extends React.Component<IOperatorHeaderProps> {
  public state = {
    modalIsOpen: false,
  };

  public render() {
    const {
      id,
      icon,
      description,
      cluster,
      namespace,
      version,
      provider,
      hideButton,
      disableButton,
    } = this.props;
    return (
      <header>
        <div className="ChartView__heading margin-normal row">
          <div className="col-1 ChartHeader__icon">
            <div className="ChartIcon">
              <img className="ChartIcon__img" src={icon} alt="icon" />
            </div>
          </div>
          <div className="col-9">
            <div className="title margin-l-small">
              <h1 className="margin-t-reset">{id}</h1>
              <h5 className="subtitle margin-b-normal">
                {/* TODO(andresmgot): Filter by provider */}
                <span>{version} - Provided by </span>
                <Link to={app.operators.list(cluster, namespace)}>{provider}</Link>
              </h5>
              <h5 className="subtitle margin-b-reset">{description}</h5>
            </div>
          </div>
          {!hideButton && (
            <div className="col-2 ChartHeader__button">
              <button
                className="button button-primary button-accent"
                onClick={this.redirect}
                disabled={disableButton}
              >
                {disableButton ? "Deployed" : "Deploy"}
              </button>
            </div>
          )}
        </div>
        <hr />
      </header>
    );
  }

  public redirect = () => {
    this.props.push(app.operators.new(this.props.cluster, this.props.namespace, this.props.id));
  };
}

export default OperatorHeader;
