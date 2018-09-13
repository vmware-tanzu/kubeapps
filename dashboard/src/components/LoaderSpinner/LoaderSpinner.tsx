import * as React from "react";

import "./LoaderSpinner.css";

const LoaderSpinner: React.SFC<{}> = _ => {
  // Based on http://tobiasahlin.com/spinkit/
  return (
    <div className="spinner">
      <div className="spinner__bounce1" />
      <div className="spinner__bounce2" />
    </div>
  );
};

export default LoaderSpinner;
