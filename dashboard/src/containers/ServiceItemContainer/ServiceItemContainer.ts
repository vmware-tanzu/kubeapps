import * as _ from "lodash";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import ServiceItem from "../../components/AppView/ServicesTable/ServiceItem";
import { Kube } from "../../shared/Kube";
import { IResourceRef, IStoreState } from "../../shared/types";

interface IServiceItemContainerProps {
  serviceRef: IResourceRef;
}

function mapStateToProps({ kube }: IStoreState, props: IServiceItemContainerProps) {
  const { serviceRef } = props;
  const serviceKey = Kube.getResourceURL(
    serviceRef.apiVersion,
    "services",
    serviceRef.namespace,
    serviceRef.name,
  );
  return {
    ...props,
    service: kube.items[serviceKey],
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  props: IServiceItemContainerProps,
) {
  const { serviceRef } = props;
  return {
    getService: () =>
      dispatch(
        actions.kube.getResource(
          serviceRef.apiVersion,
          "services",
          serviceRef.namespace,
          serviceRef.name,
        ),
      ),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ServiceItem);
