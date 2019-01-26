import * as _ from "lodash";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import SecretItem from "../../components/AppView/SecretsTable/SecretItem";
import ResourceRef from "../../shared/ResourceRef";
import { IKubeItem, ISecret, IStoreState } from "../../shared/types";

interface ISecretItemContainerProps {
  secretRef: ResourceRef;
}

function mapStateToProps({ kube }: IStoreState, props: ISecretItemContainerProps) {
  const { secretRef } = props;
  return {
    name: secretRef.name,
    // convert IKubeItem<IResource> to IKubeItem<ISecret>
    secret: kube.items[secretRef.getResourceURL()] as IKubeItem<ISecret>,
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  props: ISecretItemContainerProps,
) {
  const { secretRef } = props;
  return {
    getSecret: () =>
      dispatch(
        actions.kube.getResource(
          secretRef.apiVersion,
          "secrets",
          secretRef.namespace,
          secretRef.name,
        ),
      ),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SecretItem);
