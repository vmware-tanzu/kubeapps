import * as React from "react";

import LoadingWrapper from "../LoadingWrapper";
import PageHeader from "../PageHeader";
import OLMNotFound from "./OLMNotFound";

export interface IOperatorListProps {
  isFetching: boolean;
  checkOLMInstalled: () => Promise<boolean>;
  isOLMInstalled: boolean;
}

class OperatorList extends React.Component<IOperatorListProps> {
  public componentDidMount() {
    this.props.checkOLMInstalled();
  }

  public render() {
    const { isFetching, isOLMInstalled } = this.props;
    return (
      <div>
        <PageHeader>
          <h1>Operators</h1>
        </PageHeader>
        <main>
          <LoadingWrapper loaded={!isFetching}>
            {isOLMInstalled ? <p>OLM Installed!</p> : <OLMNotFound />}
          </LoadingWrapper>
        </main>
      </div>
    );
  }
}

export default OperatorList;
