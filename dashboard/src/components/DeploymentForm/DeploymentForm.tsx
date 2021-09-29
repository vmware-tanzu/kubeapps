import actions from "actions";
import AvailablePackageDetailExcerpt from "components/Catalog/AvailablePackageDetailExcerpt";
import ChartHeader from "components/ChartView/ChartHeader";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { push } from "connected-react-router";
import { AvailablePackageReference } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
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
  pluginName: string;
  pluginVersion: string;
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
    pluginName,
    pluginVersion,
    version: chartVersion,
  } = ReactRouter.useParams() as IRouteParams;
  const {
    apps,
    config,
    charts: { isFetching: chartsIsFetching, selected },
  } = useSelector((state: IStoreState) => state);
  const packageId = `${repo}/${id}`;
  const chartNamespace = global === "global" ? config.kubeappsNamespace : namespace;
  const chartCluster = global === "global" ? config.kubeappsCluster : cluster;
  const error = apps.error || selected.error;
  const kubeappsNamespace = config.kubeappsNamespace;
  const { availablePackageDetail, versions, schema, values, pkgVersion } = selected;
  const [isDeploying, setDeploying] = useState(false);
  const [releaseName, setReleaseName] = useState("");
  const [appValues, setAppValues] = useState(values || "");
  const [valuesModified, setValuesModified] = useState(false);
  const [pluginObj] = useState(
    selected.availablePackageDetail?.availablePackageRef?.plugin ??
      ({ name: pluginName, version: pluginVersion } as Plugin),
  );

  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersions({
        context: { cluster: chartCluster, namespace: chartNamespace },
        plugin: pluginObj,
        identifier: packageId,
      } as AvailablePackageReference),
    );
  }, [dispatch, chartCluster, chartNamespace, packageId, pluginObj]);

  useEffect(() => {
    if (!valuesModified) {
      setAppValues(values || "");
    }
  }, [values, valuesModified]);

  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersion(
        {
          context: { cluster: chartCluster, namespace: chartNamespace },
          plugin: pluginObj,
          identifier: packageId,
        } as AvailablePackageReference,
        chartVersion || versions[0]?.pkgVersion,
      ),
    );
  }, [chartCluster, chartNamespace, packageId, chartVersion, dispatch, pluginObj, versions]);

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
    if (availablePackageDetail) {
      const deployed = await dispatch(
        actions.apps.deployChart(
          cluster,
          namespace,
          availablePackageDetail,
          releaseName,
          appValues,
          schema,
        ),
      );
      setDeploying(false);
      if (deployed) {
        dispatch(
          push(
            url.app.apps.get({
              context: { cluster: cluster, namespace: namespace },
              plugin: pluginObj,
              identifier: releaseName,
            } as AvailablePackageReference),
          ),
        );
      }
    }
  };

  const selectVersion = (e: React.ChangeEvent<HTMLSelectElement>) => {
    dispatch(
      push(
        url.app.apps.new(
          cluster,
          namespace,
          availablePackageDetail!,
          e.currentTarget.value,
          kubeappsNamespace,
          pluginObj,
        ),
      ),
    );
  };

  if (error?.constructor === FetchError) {
    return (
      error && (
        <Alert theme="danger">
          Unable to retrieve the current app: {(error as FetchError).message}
        </Alert>
      )
    );
  }

  if (!availablePackageDetail) {
    return <LoadingWrapper className="margin-t-xxl" loadingText={`Fetching ${packageId}...`} />;
  }
  return (
    <section>
      <ChartHeader
        chartAttrs={availablePackageDetail}
        versions={versions}
        onSelect={selectVersion}
        selectedVersion={pkgVersion}
      />
      {isDeploying && (
        <h3 className="center" style={{ marginBottom: "1.2rem" }}>
          Hang tight, the application is being deployed...
        </h3>
      )}
      <LoadingWrapper loaded={!isDeploying}>
        <Row>
          <Column span={3}>
            <AvailablePackageDetailExcerpt pkg={availablePackageDetail} />
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
                packageId={packageId}
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
