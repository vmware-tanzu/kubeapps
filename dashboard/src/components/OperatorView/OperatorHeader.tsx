import * as React from "react";
import { Link } from "react-router-dom";
import OperatorInstallation from "./OperatorInstallation";

interface IOperatorHeaderProps {
  id: string;
  icon?: string;
  description: string;
  namespace: string;
  version: string;
  provider: string;
  namespaced: boolean;
}

class OperatorHeader extends React.Component<IOperatorHeaderProps> {
  public state = {
    modalIsOpen: false,
  };

  public render() {
    const { id, icon, description, namespace, version, provider } = this.props;
    return (
      <header>
        <div className="ChartView__heading margin-normal row">
          <div className="col-1 ChartHeader__icon">
            <div className="ChartIcon">
              <img className="ChartIcon__img" src={icon} />
            </div>
          </div>
          <div className="col-9">
            <div className="title margin-l-small">
              <h1 className="margin-t-reset">{id}</h1>
              <h5 className="subtitle margin-b-normal">
                {/* TODO(andresmgot): Filter by provider */}
                <span>{version} - Provided by </span>
                <Link to={`/ns/${namespace}/operators`}>{provider}</Link>
              </h5>
              <h5 className="subtitle margin-b-reset">{description}</h5>
            </div>
          </div>
          <div className="col-2 ChartHeader__button">
            <button className="button button-primary button-accent" onClick={this.openModal}>
              Deploy
            </button>
            <OperatorInstallation
              name={id}
              namespaced={this.props.namespaced}
              closeModal={this.closeModal}
              modalIsOpen={this.state.modalIsOpen}
            />
          </div>
        </div>
        <hr />
      </header>
    );
  }

  public openModal = () => {
    this.setState({
      modalIsOpen: true,
    });
  };

  public closeModal = () => {
    this.setState({
      modalIsOpen: false,
    });
  };
}

export default OperatorHeader;
