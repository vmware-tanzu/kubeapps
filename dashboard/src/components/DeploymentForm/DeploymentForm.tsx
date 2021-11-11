import actions from "actions";
import AvailablePackageDetailExcerpt from "components/Catalog/AvailablePackageDetailExcerpt";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Row from "components/js/Row";
import PackageHeader from "components/PackageHeader/PackageHeader";
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
  pluginName: string;
  pluginVersion: string;
  packageCluster: string;
  packageNamespace: string;
  packageId: string;
  packageVersion?: string;
}

export default function DeploymentForm() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const {
    cluster: targetCluster,
    namespace: targetNamespace,
    packageId,
    pluginName,
    pluginVersion,
    packageCluster,
    packageNamespace,
    packageVersion,
  } = ReactRouter.useParams() as IRouteParams;
  const {
    packages: { isFetching: packagesIsFetching, selected: selectedPackage },
    apps,
  } = useSelector((state: IStoreState) => state);

  const [isDeploying, setDeploying] = useState(false);
  const [releaseName, setReleaseName] = useState("");
  const [appValues, setAppValues] = useState(selectedPackage.values || "");
  const [valuesModified, setValuesModified] = useState(false);

  const error = apps.error || selectedPackage.error;

  const [pluginObj] = useState({ name: pluginName, version: pluginVersion } as Plugin);

  const [packageReference] = useState({
    context: {
      cluster: packageCluster,
      namespace: packageNamespace,
    },
    plugin: pluginObj,
    identifier: packageId,
  } as AvailablePackageReference);

  useEffect(() => {
    // Get the package details
    dispatch(
      actions.packages.fetchAndSelectAvailablePackageDetail(packageReference, packageVersion),
    );
    // Populate the rest of packages versions
    dispatch(actions.packages.fetchAvailablePackageVersions(packageReference));
    return () => {};
  }, [dispatch, packageReference, packageVersion]);

  useEffect(() => {
    if (!valuesModified) {
      setAppValues(selectedPackage.values || "");
    }
    return () => {};
  }, [selectedPackage.values, valuesModified]);

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
    if (selectedPackage.availablePackageDetail) {
      const deployed = await dispatch(
        // Installation always happen in the cluster/namespace passed in the URL
        actions.apps.installPackage(
          targetCluster,
          targetNamespace,
          selectedPackage.availablePackageDetail,
          releaseName,
          appValues,
          selectedPackage.schema,
        ),
      );
      setDeploying(false);
      if (deployed) {
        dispatch(
          push(
            // Redirect to the installed package, note that the cluster/ns are the ones passed
            // in the URL, not the ones from the package.
            url.app.apps.get({
              context: { cluster: targetCluster, namespace: targetNamespace },
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
        url.app.apps.new(targetCluster, targetNamespace, packageReference, e.currentTarget.value),
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

  if (!selectedPackage.availablePackageDetail) {
    return (
      <LoadingWrapper
        className="margin-t-xxl"
        loadingText={`Fetching ${decodeURIComponent(packageId)}...`}
      />
    );
  }
  return (
    <section>
      <PackageHeader
        availablePackageDetail={selectedPackage.availablePackageDetail}
        versions={selectedPackage.versions}
        onSelect={selectVersion}
        selectedVersion={selectedPackage.pkgVersion}
      />
      {isDeploying && (
        <h3 className="center" style={{ marginBottom: "1.2rem" }}>
          Hang tight, the application is being deployed...
        </h3>
      )}
      <LoadingWrapper loaded={!isDeploying}>
        <Row>
          <Column span={3}>
            <AvailablePackageDetailExcerpt pkg={selectedPackage.availablePackageDetail} />
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
                packageVersion={packageVersion!}
                packagesIsFetching={packagesIsFetching}
                selected={selectedPackage}
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
