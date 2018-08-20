import * as React from "react";

interface IAlphaWarningProps {
  disableWarning: () => Promise<any>;
}

class AlphaWarning extends React.Component<IAlphaWarningProps> {
  public render() {
    return (
      <div className="row">
        <div className="col-5">
          <div className="margin-b-normal margin-t-normal">
            <div className="alert alert-warning margin-c" role="alert">
              Service Catalog integration is under heavy development. If you find an issue please
              report it{" "}
              <a target="_blank" href="https://github.com/kubeapps/kubeapps/issues">
                {" "}
                here.{" "}
              </a>
              <button className="alert__close" onClick={this.props.disableWarning}>
                &times;
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default AlphaWarning;
