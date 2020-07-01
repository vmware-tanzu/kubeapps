import * as React from "react";

import "./CardGrid.v2.css";

function CardGrid(props: { children: JSX.Element }) {
  return (
    <div className="card-grid">
      <div className="clr-row">{props.children}</div>
    </div>
  );
}

export default CardGrid;
