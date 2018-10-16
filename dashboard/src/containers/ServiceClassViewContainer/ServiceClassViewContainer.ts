import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";

import { JSONSchema6 } from "json-schema";
import ServiceClassView from "../../components/ServiceClassView";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      brokerName: string;
      className: string;
    };
  };
}

function mapStateToProps({ catalog, namespace }: IStoreState, { match: { params } }: IRouteProps) {
  // TODO: Move svcClass filter to Component
  const svcClass =
    catalog.classes.list.find(potential => !!(potential.spec.externalName === params.className)) ||
    undefined;
  return {
    classes: catalog.classes,
    classname: params.className,
    createError: catalog.errors.create,
    error: catalog.errors.fetch,
    namespace: namespace.current,
    plans: catalog.plans,
    svcClass,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getClasses: async () => {
      dispatch(actions.catalog.getClasses());
    },
    getPlans: async () => {
      dispatch(actions.catalog.getPlans());
    },
    provision: (
      instanceName: string,
      namespace: string,
      className: string,
      planName: string,
      parameters: {},
      schema?: JSONSchema6,
    ) => {
      return dispatch(
        actions.catalog.provision(instanceName, namespace, className, planName, parameters, schema),
      );
    },
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ServiceClassView);
