import * as React from "react";

import "./LoadingSpinner.css";

export interface ISpinnerProps {
  size?: string;
}

const LoaderSpinner: React.SFC<ISpinnerProps> = props => {
  // Based on http://tobiasahlin.com/spinkit/
  return (
    <div className={`spinner ${props.size && `spinner--${props.size}`}`}>
      <div className="spinner__bounce1" />
      <div className="spinner__bounce2" />
    </div>
  );
};

export default LoaderSpinner;
