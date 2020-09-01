import React, { useEffect } from "react";

import actions from "actions";
import { CdsButton, CdsIcon } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported.v2";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import { push } from "connected-react-router";
import { useDispatch, useSelector } from "react-redux";
import { Operators } from "../../shared/Operators";
import { IStoreState } from "../../shared/types";
import { api, app } from "../../shared/url";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import OperatorDescription from "./OperatorDescription.v2";
import OperatorHeader from "./OperatorHeader.v2";

interface IOperatorViewProps {
  operatorName: string;
  cluster: string;
  namespace: string;
}

export default function OperatorView({ operatorName, cluster, namespace }: IOperatorViewProps) {
  const dispatch = useDispatch();
  useEffect(() => {
    dispatch(actions.operators.getOperator(cluster, namespace, operatorName));
  }, [dispatch, cluster, namespace, operatorName]);

  const {
    operators: {
      operator,
      isFetching,
      errors: {
        operator: { fetch: error },
      },
      csv,
    },
    config: { kubeappsCluster },
  } = useSelector((state: IStoreState) => state);

  useEffect(() => {
    if (operator) {
      const defaultChannel = Operators.getDefaultChannel(operator);
      if (defaultChannel) {
        dispatch(actions.operators.getCSV(cluster, namespace, defaultChannel.currentCSV));
      }
    }
  }, [dispatch, operator, cluster, namespace]);

  const redirect = () => dispatch(push(app.operators.new(cluster, namespace, operatorName)));

  if (cluster !== kubeappsCluster) {
    return <OperatorNotSupported kubeappsCluster={kubeappsCluster} namespace={namespace} />;
  }
  if (error) {
    return (
      <Alert theme="danger">
        Found an error while fetching {operatorName}: {error.message}
      </Alert>
    );
  }
  if (isFetching || !operator) {
    return <LoadingWrapper />;
  }
  const channel = Operators.getDefaultChannel(operator);
  if (!channel) {
    return (
      <Alert theme="danger">
        Operator {operatorName} doesn't define a valid channel. This is needed to extract required
        info.
      </Alert>
    );
  }
  const { currentCSVDesc } = channel;
  return (
    <section>
      <div className="header-button">
        <OperatorHeader
          title={`${operator.metadata.name} by ${operator.status.provider.name}`}
          icon={api.operators.operatorIcon(namespace, operator.metadata.name)}
          version={currentCSVDesc.version}
        >
          <CdsButton status="primary" disabled={!!csv} onClick={redirect}>
            <CdsIcon shape="deploy" inverse={true} /> Deploy
          </CdsButton>
        </OperatorHeader>
      </div>
      <section>
        <Row>
          <Column span={3}>
            <OperatorSummary />
          </Column>
          <Column span={9}>
            <OperatorDescription description={currentCSVDesc.description} />
            <div className="after-readme-button">
              <CdsButton status="primary" disabled={!!csv} onClick={redirect}>
                <CdsIcon shape="deploy" inverse={true} /> Deploy
              </CdsButton>
            </div>
          </Column>
        </Row>
      </section>
    </section>
  );
}
