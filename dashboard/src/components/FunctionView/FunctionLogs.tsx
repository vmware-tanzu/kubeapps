import * as React from "react";
import { LazyStream, LineNumber } from "react-lazylog";

import { IFunction } from "../../shared/types";
import "./FunctionLogs.css";

interface IFunctionLogsProps {
  function: IFunction;
  podName?: string;
}

class FunctionLogs extends React.Component<IFunctionLogsProps> {
  public componentDidMount() {
    LineNumber.defaultProps.style = { width: "30px", minWidth: "30px" };
  }

  public render() {
    const { function: f, podName } = this.props;
    const url =
      podName &&
      `/api/kube/api/v1/namespaces/${f.metadata.namespace}/pods/${podName}/log?follow=true`;
    return (
      <div className="FunctionLogs">
        <h6>Logs</h6>
        <hr />
        {url ? <LazyStream follow={true} url={url} /> : <div>Loading</div>}
      </div>
    );
  }
}
export default FunctionLogs;
