import * as _ from "lodash";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import SecretsTable from "../../components/AppView/SecretsTable";
import { IKubeItem, ISecret, IStoreState } from "../../shared/types";

interface ISecretTableProps {
  namespace: string;
  secretNames: string[];
}

function filterByResourceType(type: string, resources: { [s: string]: any }) {
  return _.pickBy(resources, (r, k) => {
    return k.indexOf(`/${type}/`) > -1;
  });
}

function mapStateToProps({ kube }: IStoreState, props: ISecretTableProps) {
  return {
    namespace: props.namespace,
    secretNames: props.secretNames,
    secrets: _.map(filterByResourceType("secrets", kube.items), (i: IKubeItem<ISecret>) => i),
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getSecret: (namespace: string, name: string) =>
      dispatch(actions.kube.getResource("v1", "secrets", namespace, name)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SecretsTable);
