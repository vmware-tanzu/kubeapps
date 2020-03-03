import * as React from "react";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import { IStoreState } from "../../shared/types";

function mapStateToProps(
  { apps, namespace, charts }: IStoreState,
  { location }: RouteComponentProps<{}>,
) {
  return {};
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {};
}

class Comp extends React.Component {
  public render() {
    return <h1>This component has not been built yet</h1>;
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Comp);
