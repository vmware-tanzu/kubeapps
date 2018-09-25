import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { push } from "react-router-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";

import FunctionList from "../../components/FunctionList";
import { IFunction, IStoreState } from "../../shared/types";

function mapStateToProps(
  { functions: { errors, items, runtimes }, namespace }: IStoreState,
  { location }: RouteComponentProps<{}>,
) {
  return {
    createError: errors.create,
    error: errors.fetch,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }).q || "",
    functions: items,
    namespace: namespace.current,
    runtimes,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    deployFunction: (name: string, namespace: string, spec: IFunction["spec"]) =>
      dispatch(actions.functions.createFunction(name, namespace, spec)),
    fetchFunctions: (ns: string) => dispatch(actions.functions.fetchFunctions(ns)),
    fetchRuntimes: () => dispatch(actions.functions.fetchRuntimes()),
    navigateToFunction: (name: string, namespace: string) =>
      dispatch(push(`/functions/ns/${namespace}/${name}`) as any),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter) as any),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(FunctionList);
