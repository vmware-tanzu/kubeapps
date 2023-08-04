// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import Column from "components/Column";
import LoadingWrapper from "components/LoadingWrapper";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import Row from "components/Row";
import { usePush } from "hooks/push";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useParams } from "react-router-dom";
import { Operators } from "shared/Operators";
import { IStoreState } from "shared/types";
import { api, app } from "shared/url";
import OperatorDescription from "./OperatorDescription";
import OperatorHeader from "./OperatorHeader";

export default function OperatorView() {
  const dispatch = useDispatch();
  const {
    operators: {
      operator,
      isFetching,
      errors: {
        operator: { fetch: error },
      },
      subscriptions,
    },
    clusters: { currentCluster: cluster, clusters },
  } = useSelector((state: IStoreState) => state);
  const namespace = clusters[cluster].currentNamespace;

  type IOperatorViewParams = {
    operator: string;
  };
  const { operator: operatorName } = useParams<IOperatorViewParams>();

  useEffect(() => {
    dispatch(actions.operators.getOperator(cluster, namespace, operatorName || ""));
    dispatch(actions.operators.listSubscriptions(cluster, namespace));
  }, [dispatch, cluster, namespace, operatorName]);

  useEffect(() => {
    if (operator) {
      const defaultChannel = Operators.getDefaultChannel(operator);
      if (defaultChannel) {
        dispatch(actions.operators.getCSV(cluster, namespace, defaultChannel.currentCSV));
      }
    }
  }, [dispatch, operator, cluster, namespace]);

  const push = usePush();
  const redirect = () => push(app.operators.new(cluster, namespace, operatorName || ""));

  if (error) {
    return (
      <AlertGroup status="danger">
        An error occurred while fetching the Operator {operatorName}: {error.message}.
      </AlertGroup>
    );
  }
  if (isFetching || !operator) {
    return <LoadingWrapper className="margin-t-xxl" loadingText="Fetching Operator..." />;
  }
  const channel = Operators.getDefaultChannel(operator);
  if (!channel) {
    return (
      <AlertGroup status="danger">
        Operator {operatorName} doesn't define a valid channel. This is needed to extract required
        info.
      </AlertGroup>
    );
  }
  const { currentCSVDesc } = channel;
  const alreadyInstalled = subscriptions.some(s => s.spec.name === operator.metadata.name);
  return (
    <section>
      <OperatorHeader
        title={`${operator.metadata.name} by ${operator.status.provider.name}`}
        icon={api.operators.operatorIcon(cluster, namespace, operator.metadata.name)}
        version={currentCSVDesc.version}
        buttons={[
          <CdsButton
            key="deploy-button"
            status="primary"
            disabled={alreadyInstalled}
            onClick={redirect}
          >
            <CdsIcon shape="deploy" /> Deploy
          </CdsButton>,
        ]}
      />
      <section>
        <Row>
          <Column span={3}>
            <OperatorSummary />
          </Column>
          <Column span={9}>
            <OperatorDescription description={currentCSVDesc.description} />
            <div className="after-readme-button">
              <CdsButton status="primary" disabled={alreadyInstalled} onClick={redirect}>
                <CdsIcon shape="deploy" /> Deploy
              </CdsButton>
            </div>
          </Column>
        </Row>
      </section>
    </section>
  );
}
