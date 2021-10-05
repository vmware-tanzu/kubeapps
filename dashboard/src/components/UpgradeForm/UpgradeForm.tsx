import actions from "actions";
import AvailablePackageDetailExcerpt from "components/Catalog/AvailablePackageDetailExcerpt";
import ChartHeader from "components/ChartView/ChartHeader";
import ChartVersionSelector from "components/ChartView/ChartVersionSelector";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { push } from "connected-react-router";
import * as jsonpatch from "fast-json-patch";
import {
  AvailablePackageReference,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as yaml from "js-yaml";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { deleteValue, setValue } from "../../shared/schema";
import { IChartState, IStoreState } from "../../shared/types";
import * as url from "../../shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import "./UpgradeForm.css";

export interface IUpgradeFormProps {
  appCurrentVersion: string;
  appCurrentValues?: string;
  packageId: string;
  chartsIsFetching: boolean;
  namespace: string;
  cluster: string;
  releaseName: string;
  repoNamespace: string;
  error?: Error;
  selected: IChartState["selected"];
  deployed: IChartState["deployed"];
  plugin: Plugin;
}

function applyModifications(mods: jsonpatch.Operation[], values: string) {
  // And we add any possible change made to the original version
  if (mods.length) {
    mods.forEach(modification => {
      if (modification.op === "remove") {
        values = deleteValue(values, modification.path);
      } else {
        // Transform the modification as a ReplaceOperation to read its value
        const value = (modification as jsonpatch.ReplaceOperation<any>).value;
        values = setValue(values, modification.path, value);
      }
    });
  }
  return values;
}

function UpgradeForm({
  appCurrentVersion,
  appCurrentValues,
  packageId,
  chartsIsFetching,
  namespace,
  cluster,
  releaseName,
  repoNamespace,
  error,
  selected,
  deployed,
  plugin,
}: IUpgradeFormProps) {
  const [appValues, setAppValues] = useState(appCurrentValues || "");
  const [isDeploying, setIsDeploying] = useState(false);
  const [valuesModified, setValuesModified] = useState(false);
  const [modifications, setModifications] = useState(
    undefined as undefined | jsonpatch.Operation[],
  );
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const [deployedValues, setDeployedValues] = useState("");

  const { availablePackageDetail, versions, schema, values, pkgVersion } = selected;

  const packageCluster = availablePackageDetail?.availablePackageRef?.context?.cluster;

  const {
    apps: { isFetching: appsFetching },
    charts: { isFetching: chartsFetching },
  } = useSelector((state: IStoreState) => state);
  const isFetching = appsFetching || chartsFetching;
  const pluginObj = plugin ?? selected.availablePackageDetail?.availablePackageRef?.plugin;

  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersions({
        context: {
          cluster: packageCluster ?? cluster,
          namespace: repoNamespace,
        },
        plugin: pluginObj,
        identifier: packageId,
      } as AvailablePackageReference),
    );
  }, [dispatch, packageCluster, repoNamespace, packageId, cluster, pluginObj]);

  useEffect(() => {
    if (deployed.values && !modifications) {
      // Calculate modifications from the default values
      const defaultValuesObj = yaml.load(deployed.values);
      const deployedValuesObj = yaml.load(appCurrentValues || "");
      const newModifications = jsonpatch.compare(defaultValuesObj as any, deployedValuesObj as any);
      const values = applyModifications(newModifications, deployed.values);
      setModifications(newModifications);
      setAppValues(values);
    }
  }, [deployed.values, appCurrentValues, modifications]);

  useEffect(() => {
    if (deployed.values) {
      // Apply modifications to deployed values
      const values = applyModifications(modifications || [], deployed.values);
      setDeployedValues(values);
    }
  }, [deployed.values, modifications]);

  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersion(
        {
          context: { cluster: packageCluster, namespace: repoNamespace },
          plugin: pluginObj,
          identifier: packageId,
        } as AvailablePackageReference,
        deployed.chartVersion?.version?.pkgVersion,
      ),
    );
  }, [
    dispatch,
    packageCluster,
    repoNamespace,
    packageId,
    deployed.chartVersion?.version?.pkgVersion,
    pluginObj,
  ]);

  useEffect(() => {
    if (!valuesModified && values) {
      // Apply modifications to the new selected version
      const newAppValues = modifications?.length
        ? applyModifications(modifications, values)
        : values;
      setAppValues(newAppValues);
    }
  }, [values, modifications, valuesModified]);

  const setValuesModifiedTrue = () => {
    setValuesModified(true);
  };

  const handleValuesChange = (value: string) => {
    setAppValues(value);
  };

  const selectVersion = (e: React.ChangeEvent<HTMLSelectElement>) => {
    dispatch(
      actions.charts.fetchChartVersion(
        {
          context: { cluster: packageCluster, namespace: repoNamespace },
          plugin: pluginObj,
          identifier: packageId,
        } as AvailablePackageReference,
        e.currentTarget.value,
      ),
    );
  };

  const handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsDeploying(true);
    if (availablePackageDetail) {
      const deployedSuccess = await dispatch(
        actions.apps.upgradeApp(
          {
            context: { cluster: cluster, namespace: namespace },
            identifier: releaseName,
            plugin: pluginObj,
          } as InstalledPackageReference,
          availablePackageDetail,
          repoNamespace,
          appValues,
          schema,
        ),
      );
      setIsDeploying(false);
      if (deployedSuccess) {
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

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <section>
      {isDeploying && (
        <h3 className="center" style={{ marginBottom: "1.2rem" }}>
          The application is being upgraded, please wait...
        </h3>
      )}
      {!isFetching && error && <Alert theme="danger">An error occurred: {error?.message}</Alert>}
      <LoadingWrapper
        loaded={!isDeploying && !isFetching && versions?.length > 0 && !!availablePackageDetail}
      >
        {(!isFetching && versions?.length === 0) || !availablePackageDetail ? (
          <></>
        ) : (
          <>
            <ChartHeader
              releaseName={releaseName}
              chartAttrs={availablePackageDetail}
              versions={versions}
              onSelect={selectVersion}
              currentVersion={deployed.chartVersion?.version?.pkgVersion}
              selectedVersion={pkgVersion}
            />
            <LoadingWrapper
              loaded={
                !isDeploying && !isFetching && versions?.length > 0 && !!availablePackageDetail
              }
            >
              <Row>
                <Column span={3}>
                  <AvailablePackageDetailExcerpt pkg={availablePackageDetail} />
                </Column>
                <Column span={9}>
                  <form onSubmit={handleDeploy}>
                    <div className="upgrade-form-version-selector">
                      <label className="centered deployment-form-label deployment-form-label-text-param">
                        Upgrade to Version
                      </label>
                      <ChartVersionSelector
                        versions={versions}
                        selectedVersion={pkgVersion}
                        onSelect={selectVersion}
                        currentVersion={appCurrentVersion}
                        chartAttrs={availablePackageDetail}
                      />
                    </div>
                    <DeploymentFormBody
                      deploymentEvent="upgrade"
                      packageId={packageId}
                      chartVersion={appCurrentVersion}
                      deployedValues={deployedValues}
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
          </>
        )}
      </LoadingWrapper>
    </section>
  );
}

export default UpgradeForm;
