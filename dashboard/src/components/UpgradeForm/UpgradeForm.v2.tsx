import { RouterAction } from "connected-react-router";
import * as jsonpatch from "fast-json-patch";
import { JSONSchema4 } from "json-schema";
import React, { useEffect, useState } from "react";
import * as YAML from "yaml";

import ChartSummary from "components/Catalog/ChartSummary";
import ChartHeader from "components/ChartView/ChartHeader.v2";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { deleteValue, setValue } from "../../shared/schema";
import { IChartState, IChartVersion } from "../../shared/types";
import * as url from "../../shared/url";
import DeploymentFormBody from "../DeploymentFormBody/DeploymentFormBody.v2";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";

export interface IUpgradeFormProps {
  appCurrentVersion: string;
  appCurrentValues?: string;
  chartName: string;
  chartsIsFetching: boolean;
  namespace: string;
  cluster: string;
  releaseName: string;
  repo: string;
  repoNamespace: string;
  error?: Error;
  selected: IChartState["selected"];
  deployed: IChartState["deployed"];
  upgradeApp: (
    cluster: string,
    namespace: string,
    version: IChartVersion,
    chartNamespace: string,
    releaseName: string,
    values?: string,
    schema?: JSONSchema4,
  ) => Promise<boolean>;
  push: (location: string) => RouterAction;
  fetchChartVersions: (namespace: string, id: string) => Promise<IChartVersion[]>;
  getChartVersion: (namespace: string, id: string, chartVersion: string) => void;
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
  chartName,
  chartsIsFetching,
  namespace,
  cluster,
  releaseName,
  repo,
  repoNamespace,
  error,
  selected,
  deployed,
  upgradeApp,
  push,
  fetchChartVersions,
  getChartVersion,
}: IUpgradeFormProps) {
  const [appValues, setAppValues] = useState(appCurrentValues || "");
  const [isDeploying, setIsDeploying] = useState(false);
  const [valuesModified, setValuesModified] = useState(false);
  const [modifications, setModifications] = useState(
    undefined as undefined | jsonpatch.Operation[],
  );
  const [deployedValues, setDeployedValues] = useState("");

  const chartID = `${repo}/${chartName}`;
  const { version } = selected;

  useEffect(() => {
    fetchChartVersions(repoNamespace, chartID);
  }, [fetchChartVersions, repoNamespace, chartID]);

  useEffect(() => {
    if (deployed.values && !modifications) {
      // Calculate modifications from the default values
      const defaultValuesObj = YAML.parse(deployed.values);
      const deployedValuesObj = YAML.parse(appCurrentValues || "");
      const newModifications = jsonpatch.compare(defaultValuesObj, deployedValuesObj);
      const values = applyModifications(newModifications, deployed.values);
      setModifications(newModifications);
      setAppValues(values);
      setDeployedValues(values);
    }
  }, [deployed.values, appCurrentValues, modifications]);

  useEffect(() => {
    if (deployed.chartVersion?.attributes.version) {
      getChartVersion(repoNamespace, chartID, deployed.chartVersion?.attributes.version);
    }
  }, [getChartVersion, repoNamespace, chartID, deployed.chartVersion]);

  useEffect(() => {
    if (!valuesModified && selected.values) {
      // Apply modifications to the new selected version
      const newAppValues = modifications?.length
        ? applyModifications(modifications, selected.values)
        : selected.values;
      setAppValues(newAppValues);
    }
  }, [selected.values, modifications, valuesModified]);

  const setValuesModifiedTrue = () => {
    setValuesModified(true);
  };

  const handleValuesChange = (value: string) => {
    setAppValues(value);
  };

  const selectVersion = (e: React.ChangeEvent<HTMLSelectElement>) => {
    getChartVersion(repoNamespace, chartID, e.currentTarget.value);
  };

  const handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsDeploying(true);
    if (selected.version) {
      const deployedSuccess = await upgradeApp(
        cluster,
        namespace,
        selected.version,
        repoNamespace,
        releaseName,
        appValues,
        selected.schema,
      );
      setIsDeploying(false);
      if (deployedSuccess) {
        push(url.app.apps.get(cluster, namespace, releaseName));
      }
    }
  };

  if (error) {
    return <Alert theme="danger">Found error: {error.message}</Alert>;
  }
  if (selected.versions.length === 0 || !version) {
    return <LoadingWrapper loaded={false} />;
  }

  const chartAttrs = version.relationships.chart.data;
  return (
    <section>
      <ChartHeader
        releaseName={releaseName}
        chartAttrs={chartAttrs}
        versions={selected.versions}
        onSelect={selectVersion}
        currentVersion={deployed.chartVersion?.attributes.version}
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
            <form onSubmit={handleDeploy}>
              <DeploymentFormBody
                deploymentEvent="upgrade"
                chartID={chartID}
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
    </section>
  );
}

export default UpgradeForm;
