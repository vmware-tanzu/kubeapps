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
  const { classes, errors } = catalog;

  return {
    classes,
    error: errors.fetch,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    getClasses: async () => {
      const classes = await dispatch(actions.catalog.getClasses());
      return classes;
    },
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ClassList);
