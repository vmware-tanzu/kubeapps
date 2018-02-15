import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../actions";
import { ClassList } from "../components/ClassList";
import { IStoreState } from "../shared/types";

interface IRouteProps {
  match: {
    params: {
      brokerName: string;
      className: string;
    };
  };
}

function mapStateToProps({ catalog }: IStoreState, props: IRouteProps) {
  const classes = catalog.classes;

  return {
    classes,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    getBrokers: async () => {
      const brokers = await dispatch(actions.catalog.getBrokers());
      return brokers;
    },
    getClasses: async () => {
      const classes = await dispatch(actions.catalog.getClasses());
      return classes;
    },
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ClassList);
