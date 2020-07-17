import * as React from "react";
import Row from "../js/Row";

import "./CardGrid.v2.css";

function CardGrid(props: { children: JSX.Element | JSX.Element[] }) {
  return (
    <div className="card-grid">
      <Row>{props.children}</Row>
    </div>
  );
}

export default CardGrid;
