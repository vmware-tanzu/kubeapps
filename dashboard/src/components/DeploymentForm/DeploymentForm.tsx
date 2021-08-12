import actions from "actions";
import ChartSummary from "components/Catalog/ChartSummary";
import ChartHeader from "components/ChartView/ChartHeader";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { push } from "connected-react-router";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router";
import "react-tabs/style/react-tabs.css";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { FetchError, IStoreState } from "shared/types";
import * as url from "shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";

interface IRouteParams {
  cluster: string;
  namespace: string;
  repo: string;
  global: string;
  id: string;
  version?: any;
}

export default function DeploymentForm() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const {
    cluster,
    namespace,
    repo,
    global,
    id,
    version: chartVersion,
  } = ReactRouter.useParams() as IRouteParams;
  const {
    apps,
    config,
    charts: { isFetching: chartsIsFetching, selected },
  } = useSelector((state: IStoreState) => state);
  const chartID = `${repo}/${id}`;
  const chartNamespace = global === "global" ? config.kubeappsNamespace : namespace;
  const error = apps.error || selected.error;
  const kubeappsNamespace = config.kubeappsNamespace;
  const [isDeploying, setDeploying] = useState(false);
  const [releaseName, setReleaseName] = useState("");
  const [appValues, setAppValues] = useState(selected.values || "");
  const [valuesModified, setValuesModified] = useState(false);
  const { version } = selected;

  useEffect(() => {
    dispatch(actions.charts.fetchChartVersions(cluster, chartNamespace, chartID));
  }, [dispatch, cluster, chartNamespace, chartID]);

  useEffect(() => {
    if (!valuesModified) {
      setAppValues(selected.values || "");
    }
  }, [selected.values, valuesModified]);
  useEffect(() => {
    dispatch(actions.charts.getChartVersion(cluster, chartNamespace, chartID, chartVersion));
  }, [cluster, chartNamespace, chartID, chartVersion, dispatch]);

  const handleValuesChange = (value: string) => {
    setAppValues(value);
  };

  const setValuesModifiedTrue = () => {
    setValuesModified(true);
  };

  const handleReleaseNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setReleaseName(e.target.value);
  };

  const handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setDeploying(true);
    if (selected.version) {
      const deployed = await dispatch(
        actions.apps.deployChart(
          cluster,
          namespace,
          selected.version,
          chartNamespace,
          releaseName,
          appValues,
          selected.schema,
        ),
      );
      setDeploying(false);
      if (deployed) {
        push(url.app.apps.get(cluster, namespace, releaseName));
      }
    }
  };

  const selectVersion = (e: React.ChangeEvent<HTMLSelectElement>) => {
    push(
      url.app.apps.new(
        cluster,
        namespace,
        selected.version!,
        e.currentTarget.value,
        kubeappsNamespace,
      ),
    );
  };

  if (error?.constructor === FetchError) {
    return (
      error && <Alert theme="danger">Unable to retrieve the current app: {error.message}</Alert>
    );
  }

  if (!version) {
    return <LoadingWrapper className="margin-t-xxl" loadingText={`Fetching ${chartID}...`} />;
  }
  const chartAttrs = version.relationships.chart.data;
  return (
    <section>
      <ChartHeader
        chartAttrs={chartAttrs}
        versions={selected.versions}
        onSelect={selectVersion}
        selectedVersion={selected.version?.attributes.version}
      />
      {isDeploying && (
        <h3 className="center" style={{ marginBottom: "1.2rem" }}>
          Hang tight, the application is being deployed...
        </h3>
      )}
      <LoadingWrapper loaded={!isDeploying}>
        <Row>
          <Column span={3}>
            <ChartSummary version={version} chartAttrs={chartAttrs} />
          </Column>
          <Column span={9}>
            {error && <Alert theme="danger">An error occurred: {error.message}</Alert>}
            <form onSubmit={handleDeploy}>
              <div>
                <label
                  htmlFor="releaseName"
                  className="deployment-form-label deployment-form-label-text-param"
                >
                  Name
                </label>
                <input
                  id="releaseName"
                  pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
                  title="Use lower case alphanumeric characters, '-' or '.'"
                  className="clr-input deployment-form-text-input"
                  onChange={handleReleaseNameChange}
                  value={releaseName}
                  required={true}
                />
              </div>
              <DeploymentFormBody
                deploymentEvent="install"
                chartID={chartID}
                chartVersion={chartVersion}
                chartsIsFetching={chartsIsFetching}
                selected={selected}
                setValues={handleValuesChange}
                appValues={appValues}
                setValuesModified={setValuesModifiedTrue}
              />
            </form>
          </Column>
        </Row>
      </LoadingWrapper>
    </section>
  );
}
