// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsFormGroup } from "@cds/react/forms";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import AvailablePackageDetailExcerpt from "components/Catalog/AvailablePackageDetailExcerpt";
import Column from "components/Column";
import DeploymentFormBody from "components/DeploymentForm/DeploymentFormBody";
import LoadingWrapper from "components/LoadingWrapper";
import PackageHeader from "components/PackageHeader/PackageHeader";
import PackageVersionSelector from "components/PackageHeader/PackageVersionSelector";
import Row from "components/Row";
import * as jsonpatch from "fast-json-patch";
import { usePush } from "hooks/push";
import { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import * as url from "shared/url";
import { deleteValue, parseToJS, setValue } from "shared/yamlUtils";

export interface IUpgradeFormProps {
  version?: string;
}

// TODO(agamez): Use the YAML-node based functions to avoid re-parse the yaml again and again
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

function UpgradeForm(props: IUpgradeFormProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const {
    apps: {
      selected: installedAppInstalledPackageDetail,
      isFetching: appsIsFetching,
      error,
      selectedDetails: installedAppAvailablePackageDetail,
    },
    packages: { isFetching: chartsIsFetching, selected: selectedPackage },
  } = useSelector((state: IStoreState) => state);

  const isFetching = appsIsFetching || chartsIsFetching;
  const { availablePackageDetail, versions, schema, values, pkgVersion } = selectedPackage;

  const [appValues, setAppValues] = useState("");
  const [modifications, setModifications] = useState(
    undefined as undefined | jsonpatch.Operation[],
  );
  const [deployedValues, setDeployedValues] = useState("");
  const [isDeploying, setIsDeploying] = useState(false);
  const [valuesModified, setValuesModified] = useState(false);
  const formRef = useRef<HTMLFormElement>(null);

  useEffect(() => {
    // This block just will be run once, given that populating
    // the list of versions does not depend on anything else
    if (selectedPackage.versions.length === 0) {
      dispatch(
        actions.availablepackages.fetchAvailablePackageVersions(
          installedAppInstalledPackageDetail?.availablePackageRef,
        ),
      );
      if (installedAppAvailablePackageDetail) {
        // Additionally, mark the current installed package version as the selected,
        // next time, the selection will be handled by selectVersion()
        dispatch(
          actions.availablepackages.receiveSelectedAvailablePackageDetail(
            installedAppAvailablePackageDetail,
          ),
        );
      }
      // If a version has been manually selected (eg. in the URL), fetch it explicitly
      if (props.version) {
        dispatch(
          actions.availablepackages.fetchAndSelectAvailablePackageDetail(
            installedAppInstalledPackageDetail?.availablePackageRef,
            props.version,
          ),
        );
      }
    }
  }, [
    dispatch,
    installedAppInstalledPackageDetail?.availablePackageRef,
    selectedPackage.versions.length,
    installedAppAvailablePackageDetail,
    props.version,
  ]);

  useEffect(() => {
    if (
      installedAppAvailablePackageDetail?.defaultValues &&
      installedAppInstalledPackageDetail?.valuesApplied &&
      installedAppInstalledPackageDetail?.valuesApplied !== "null" &&
      !modifications
    ) {
      // Calculate modifications from the default values
      const defaultValuesObj = parseToJS(installedAppAvailablePackageDetail?.defaultValues);
      const deployedValuesObj = parseToJS(installedAppInstalledPackageDetail?.valuesApplied);
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

  const setValuesModifiedTrue = () => {
    setValuesModified(true);
  };

  const handleValuesChange = (value: string) => {
    setAppValues(value);
  };

  const selectVersion = (e: React.ChangeEvent<HTMLSelectElement>) => {
    dispatch(
      actions.availablepackages.fetchAndSelectAvailablePackageDetail(
        installedAppInstalledPackageDetail?.availablePackageRef,
        e.currentTarget.value,
      ),
    );
  };

  const push = usePush();

  const handleDeploy = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsDeploying(true);
    if (
      availablePackageDetail &&
      installedAppInstalledPackageDetail?.installedPackageRef &&
      installedAppInstalledPackageDetail?.availablePackageRef?.context?.namespace
    ) {
      const deployedSuccess = await dispatch(
        actions.installedpackages.updateInstalledPackage(
          installedAppInstalledPackageDetail?.installedPackageRef,
          availablePackageDetail,
          appValues,
          schema,
        ),
      );
      setIsDeploying(false);
      if (deployedSuccess) {
        push(url.app.apps.get(installedAppInstalledPackageDetail?.installedPackageRef));
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
      {!isFetching && error && (
        <AlertGroup status="danger">An error occurred: {error?.message}.</AlertGroup>
      )}
      <LoadingWrapper
        loaded={!isDeploying && !isFetching && versions?.length > 0 && !!availablePackageDetail}
      >
        {(!isFetching && versions?.length === 0) || !availablePackageDetail ? (
          <></>
        ) : (
          <>
            <PackageHeader
              releaseName={installedAppInstalledPackageDetail?.installedPackageRef?.identifier}
              availablePackageDetail={availablePackageDetail}
              versions={versions}
              onSelect={selectVersion}
              currentVersion={installedAppAvailablePackageDetail?.version?.pkgVersion}
              selectedVersion={pkgVersion}
              hideVersionsSelector={true}
            />
            <LoadingWrapper
              loaded={
                !isDeploying && !isFetching && versions?.length > 0 && !!availablePackageDetail
              }
            >
              {!installedAppInstalledPackageDetail?.availablePackageRef?.identifier ||
              !installedAppInstalledPackageDetail?.currentVersion?.pkgVersion ? (
                <></>
              ) : (
                <>
                  <Row>
                    <Column span={3}>
                      <AvailablePackageDetailExcerpt pkg={availablePackageDetail} />
                    </Column>
                    <Column span={9}>
                      <form onSubmit={handleDeploy} ref={formRef}>
                        <CdsFormGroup
                          className="deployment-form"
                          layout="vertical"
                          controlWidth="shrink"
                        >
                          <PackageVersionSelector
                            versions={versions}
                            selectedVersion={pkgVersion}
                            onSelect={selectVersion}
                            currentVersion={
                              installedAppInstalledPackageDetail?.currentVersion?.pkgVersion
                            }
                            label={"Package Version"}
                            message={"Select the version this package will be upgraded to."}
                          />
                        </CdsFormGroup>
                        <DeploymentFormBody
                          deploymentEvent="upgrade"
                          packageId={
                            installedAppInstalledPackageDetail?.availablePackageRef?.identifier
                          }
                          packageVersion={
                            installedAppInstalledPackageDetail?.currentVersion?.pkgVersion
                          }
                          deployedValues={deployedValues}
                          packagesIsFetching={isFetching}
                          selected={selectedPackage}
                          setValues={handleValuesChange}
                          appValues={appValues}
                          setValuesModified={setValuesModifiedTrue}
                          formRef={formRef}
                        />
                      </form>
                    </Column>
                  </Row>
                </>
              )}
            </LoadingWrapper>
          </>
        )}
      </LoadingWrapper>
    </section>
  );
}

export default UpgradeForm;
