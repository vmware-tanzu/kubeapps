import { connect } from "react-redux";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../actions";
import { ClassView } from "../components/ClassView";
import { IStoreState } from "../shared/types";

interface IRouteProps {
  match: {
    params: {
      brokerName: string;
      className: string;
    };
  };
}

function mapStateToProps({ catalog }: IStoreState, { match: { params } }: IRouteProps) {
  const svcClass =
    catalog.classes.find(potential => !!(potential.spec.externalName === params.className)) ||
    undefined;
  return {
    classes: catalog.classes,
    classname: params.className,
    plans: catalog.plans,
    svcClass,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    getClasses: async () => {
      dispatch(actions.catalog.getClasses());
    },
    getPlans: async () => {
      dispatch(actions.catalog.getPlans());
    },
    provision: async (
      instanceName: string,
      namespace: string,
      className: string,
      planName: string,
      parameters: {},
    ) => {
      await dispatch(
        actions.catalog.provision(instanceName, namespace, className, planName, parameters),
      );
      return dispatch(actions.catalog.getCatalog(namespace));
    },
    push: (location: string) => dispatch(push(location)),
  };
}

export const ClassViewContainer = connect(mapStateToProps, mapDispatchToProps)(ClassView);
