import * as React from "react";

const TableRenderer: React.FunctionComponent<{}> = (props: any) => {
  return <table className="table">{props.children}</table>;
};

export default TableRenderer;
