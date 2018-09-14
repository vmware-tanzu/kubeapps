import * as React from "react";
import logo404 from "../../img/404.svg";

class NotFound extends React.Component {
  public render() {
    return (
      <div className="text-c align-center margin-t-huge">
        <h3>The page you are looking for can't be found.</h3>
        <img src={logo404} alt="Kubeapps logo" />
      </div>
    );
  }
}

export default NotFound;
