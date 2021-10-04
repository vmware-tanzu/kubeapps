/* eslint-disable @typescript-eslint/no-non-null-asserted-optional-chain */

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
  AvailablePackageDetail,
  AvailablePackageReference,
  InstalledPackageDetail,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
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
  installedAppAvailablePackageDetail?: AvailablePackageDetail;
  installedAppInstalledPackageDetail?: InstalledPackageDetail;
  chartsIsFetching: boolean;
  error?: Error;
  selected: IChartState["selected"];
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
  installedAppAvailablePackageDetail,
  installedAppInstalledPackageDetail,
  chartsIsFetching,
  error,
  selected,
}: IUpgradeFormProps) {
  const [appValues, setAppValues] = useState(
    installedAppInstalledPackageDetail?.valuesApplied || "",
  );
  const [isDeploying, setIsDeploying] = useState(false);
  const [valuesModified, setValuesModified] = useState(false);
  const [modifications, setModifications] = useState(
    undefined as undefined | jsonpatch.Operation[],
  );
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const [deployedValues, setDeployedValues] = useState("");
  const [hasSelectedInstalledPackage, setHasSelectedInstalledPackage] = useState(false);

  const { availablePackageDetail, versions, schema, values, pkgVersion } = selected;

  const {
    apps: { isFetching: appsFetching },
    charts: { isFetching: chartsFetching },
  } = useSelector((state: IStoreState) => state);
  const isFetching = appsFetching || chartsFetching;

  useEffect(() => {
    dispatch(
      actions.charts.fetchChartVersions({
        context: {
          cluster: installedAppInstalledPackageDetail?.installedPackageRef?.context?.cluster,
          namespace: installedAppInstalledPackageDetail?.availablePackageRef?.context?.namespace,
        },
        plugin: installedAppInstalledPackageDetail?.availablePackageRef?.plugin,
        identifier: installedAppInstalledPackageDetail?.availablePackageRef?.identifier,
      } as AvailablePackageReference),
    );
  }, [dispatch, installedAppInstalledPackageDetail]);

  useEffect(() => {
    if (installedAppAvailablePackageDetail?.defaultValues && !modifications) {
      // Calculate modifications from the default values
      const defaultValuesObj = yaml.load(installedAppAvailablePackageDetail?.defaultValues);
      const deployedValuesObj = yaml.load(installedAppInstalledPackageDetail?.valuesApplied || "");
      const newModifications = jsonpatch.compare(defaultValuesObj as any, deployedValuesObj as any);
      const values = applyModifications(
        newModifications,
        installedAppAvailablePackageDetail?.defaultValues,
      );
      setModifications(newModifications);
      setAppValues(values);
    }
  }, [
    installedAppAvailablePackageDetail?.defaultValues,
    installedAppInstalledPackageDetail?.valuesApplied,
    modifications,
  ]);

  useEffect(() => {
    if (installedAppAvailablePackageDetail?.defaultValues) {
      // Apply modifications to deployed values
      const values = applyModifications(
        modifications || [],
        installedAppAvailablePackageDetail?.defaultValues,
      );
      setDeployedValues(values);
    }
  }, [installedAppAvailablePackageDetail?.defaultValues, modifications]);

  useEffect(() => {
    if (!valuesModified && values) {
      // Apply modifications to the new selected version
      const newAppValues = modifications?.length
        ? applyModifications(modifications, values)
        : values;
      setAppValues(newAppValues);
    }
  }, [values, modifications, valuesModified]);

  // Mark the current installed package version as selected, next time,
  // the selection will be handled by selectVersion()
  useEffect(() => {
    if (!hasSelectedInstalledPackage && installedAppAvailablePackageDetail) {
      dispatch(actions.charts.selectChartVersion(installedAppAvailablePackageDetail));
      setHasSelectedInstalledPackage(true);
    }
  }, [dispatch, hasSelectedInstalledPackage, installedAppAvailablePackageDetail]);

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
          context: {
            cluster: installedAppInstalledPackageDetail?.installedPackageRef?.context?.cluster,
            namespace: installedAppInstalledPackageDetail?.availablePackageRef?.context?.namespace,
          },
          plugin: installedAppInstalledPackageDetail?.availablePackageRef?.plugin,
          identifier: installedAppInstalledPackageDetail?.availablePackageRef?.identifier,
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
            context: {
              cluster: installedAppInstalledPackageDetail?.installedPackageRef?.context?.cluster,
              namespace:
                installedAppInstalledPackageDetail?.installedPackageRef?.context?.namespace,
            },
            identifier: installedAppInstalledPackageDetail?.installedPackageRef?.identifier,
            plugin: installedAppInstalledPackageDetail?.availablePackageRef?.plugin,
          } as InstalledPackageReference,
          availablePackageDetail,
          installedAppInstalledPackageDetail?.availablePackageRef?.context?.namespace!,
          appValues,
          schema,
        ),
      );
      setIsDeploying(false);
      if (deployedSuccess) {
        dispatch(
          push(
            url.app.apps.get({
              context: {
                cluster: installedAppInstalledPackageDetail?.installedPackageRef?.context?.cluster,
                namespace:
                  installedAppInstalledPackageDetail?.installedPackageRef?.context?.namespace,
              },
              plugin: installedAppInstalledPackageDetail?.availablePackageRef?.plugin,
              identifier: installedAppInstalledPackageDetail?.installedPackageRef?.identifier,
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
              releaseName={installedAppInstalledPackageDetail?.installedPackageRef?.identifier}
              chartAttrs={availablePackageDetail}
              versions={versions}
              onSelect={selectVersion}
              currentVersion={installedAppAvailablePackageDetail?.version?.pkgVersion}
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
                        currentVersion={
                          installedAppInstalledPackageDetail?.currentVersion?.pkgVersion
                        }
                        chartAttrs={availablePackageDetail}
                      />
                    </div>
                    <DeploymentFormBody
                      deploymentEvent="upgrade"
                      packageId={
                        installedAppInstalledPackageDetail?.availablePackageRef?.identifier!
                      }
                      chartVersion={installedAppInstalledPackageDetail?.currentVersion?.pkgVersion!}
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
