// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import OperatorSummary from "components/OperatorSummary/OperatorSummary";
import { push } from "connected-react-router";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Operators } from "shared/Operators";
import { IStoreState } from "shared/types";
import { api, app } from "shared/url";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import OperatorDescription from "./OperatorDescription";
import OperatorHeader from "./OperatorHeader";

interface IOperatorViewProps {
  operatorName: string;
  cluster: string;
  namespace: string;
}

export default function OperatorView({ operatorName, cluster, namespace }: IOperatorViewProps) {
  const dispatch = useDispatch();
  useEffect(() => {
    dispatch(actions.operators.getOperator(cluster, namespace, operatorName));
    dispatch(actions.operators.listSubscriptions(cluster, namespace));
  }, [dispatch, cluster, namespace, operatorName]);

  const {
    operators: {
      operator,
      isFetching,
      errors: {
        operator: { fetch: error },
      },
      subscriptions,
    },
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

  if (error) {
    return (
      <Alert theme="danger">
        An error occurred while fetching the Operator {operatorName}: {error.message}
      </Alert>
    );
  }
  if (isFetching || !operator) {
    return <LoadingWrapper className="margin-t-xxl" loadingText="Fetching Operator..." />;
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
