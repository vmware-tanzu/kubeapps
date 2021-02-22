import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import Modal from "components/js/Modal/Modal";
import React, { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import useOutsideClick from "../js/hooks/useOutsideClick/useOutsideClick";
import Row from "../js/Row";
import "./ContextSelector.css";

function ContextSelector() {
  const location = ReactRouter.useLocation();
  const history = ReactRouter.useHistory();
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const { clusters } = useSelector((state: IStoreState) => state);
  const currentCluster = clusters.clusters[clusters.currentCluster];
  const namespaceSelected = currentCluster.currentNamespace;
  const canCreateNS = currentCluster.canCreateNS;
  const error = currentCluster.error;
  const [open, setOpen] = useState(false);
  const [cluster, setStateCluster] = useState(clusters.currentCluster);
  const [namespace, setStateNamespace] = useState(namespaceSelected);
  const [newNSModalIsOpen, setNewNSModalIsOpen] = useState(false);
  const [newNS, setNewNS] = useState("");

  // Control when users click outside
  const ref = useRef(null);
  useOutsideClick(setOpen, [ref], open);

  useEffect(() => {
    if (namespaceSelected) {
      dispatch(actions.namespace.getNamespace(clusters.currentCluster, namespaceSelected));
    }
  }, [dispatch, namespaceSelected, clusters.currentCluster]);

  useEffect(() => {
    setStateNamespace(namespaceSelected);
  }, [namespaceSelected]);

  useEffect(() => {
    setStateNamespace(clusters.clusters[cluster].currentNamespace);
  }, [clusters.clusters, cluster]);

  const toggleOpen = () => setOpen(!open);
  const selectCluster = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setStateCluster(event.target.value);
    dispatch(actions.namespace.fetchNamespaces(event.target.value));
  };
  const selectNamespace = (event: React.ChangeEvent<HTMLSelectElement>) =>
    setStateNamespace(event.target.value);
  const changeContext = () => {
    dispatch(actions.namespace.setNamespace(cluster, namespace));
    // Regex matching a namespaced route: e.g. /c/cluster/ns/namespace
    const nsRegex = /^\/c\/([^/]*)\/ns\/[^/]*\//;
    if (nsRegex.test(location.pathname)) {
      // Change the namespace in the route
      history.push(
        location.pathname
          .replace(nsRegex, `/c/${cluster}/ns/${namespace}/`)
          .concat(location.search),
      );
    }
    setOpen(false);
  };
  const openNewNSModal = () => setNewNSModalIsOpen(true);
  const closeNewNSModal = () => setNewNSModalIsOpen(false);
  const onChangeNewNS = (event: React.ChangeEvent<HTMLInputElement>) =>
    setNewNS(event.target.value);
  const createNewNS = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const created = await dispatch(actions.namespace.createNamespace(cluster, newNS));
    if (created) {
      closeNewNSModal();
      dispatch(actions.namespace.setNamespace(cluster, newNS));
      history.push(app.apps.list(cluster, newNS));
      setOpen(false);
    }
  };

  return (
    <>
      <div
        className={`dropdown kubeapps-align-center kubeapps-dropdown ${open ? "open" : ""}`}
        ref={ref}
      >
        <button
          className="kubeapps-nav-link"
          onClick={toggleOpen}
          aria-expanded={open}
          aria-haspopup="menu"
        >
          <Row>
            <Column span={10}>
              <div className="kubeapps-dropdown-section">
                <span className="kubeapps-dropdown-header">Current Context</span>
                <div>
                  <CdsIcon size="sm" shape="cluster" inverse={true} />
                  <label htmlFor="clusters" className="kubeapps-dropdown-text">
                    {clusters.currentCluster}
                  </label>
                  <CdsIcon size="sm" shape="file-group" inverse={true} />
                  <label htmlFor="namespaces" className="kubeapps-dropdown-text">
                    {namespaceSelected}
                  </label>
                </div>
              </div>
            </Column>
            <Column span={2}>
              <div
                className={`kubeapps-align-center angle ${open ? "angle-opened" : "angle-closed"}`}
              >
                <CdsIcon shape="angle" inverse={true} direction={open ? "up" : "down"} />
              </div>
            </Column>
          </Row>
        </button>
        <div className="dropdown-menu" role="menu" hidden={!open}>
          <span className="context-selector-header dropdown-menu-padding">
            Select a cluster and a namespace to manage applications
          </span>
          <div className="dropdown-menu-padding" role="menuitem">
            <CdsIcon size="sm" shape="cluster" inverse={true} />
            <span className="kubeapps-dropdown-text">Cluster</span>
            <div className="clr-select-wrapper">
              <select
                name="clusters"
                className="clr-page-size-select"
                onChange={selectCluster}
                value={cluster}
              >
                {Object.keys(clusters.clusters).map(c => {
                  return (
                    <option key={`kubeapps-dropdown-cluster-${c}`} value={c}>
                      {c}
                    </option>
                  );
                })}
              </select>
            </div>
          </div>
          <div className="dropdown-menu-padding" role="menuitem">
            <CdsIcon size="sm" shape="file-group" inverse={true} />
            <span className="kubeapps-dropdown-text">Namespace</span>
            <div className="clr-select-wrapper">
              <select
                name="namespaces"
                className="clr-page-size-select"
                onChange={selectNamespace}
                value={namespace}
              >
                {clusters.clusters[cluster].namespaces.map(n => {
                  return (
                    <option key={`kubeapps-dropdown-namespace-${n}`} value={n}>
                      {n}
                    </option>
                  );
                })}
              </select>
            </div>
            <div className="kubeapps-create-new-ns">
              <Modal
                showModal={newNSModalIsOpen}
                onModalClose={closeNewNSModal}
                title="Create a New Namespace"
              >
                <div className="newns-modal">
                  {error && <Alert theme="danger">An error occurred: {error.error.message}</Alert>}
                  <form onSubmit={createNewNS}>
                    <div className="clr-form-control">
                      <label htmlFor="namespace-name" className="clr-control-label">
                        Namespace name
                      </label>
                      <div className="clr-control-container">
                        <div className="clr-input-wrapper">
                          <input
                            type="text"
                            className="clr-input"
                            placeholder="my-namespace"
                            onChange={onChangeNewNS}
                            required={true}
                            pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
                            title="Use lower case alphanumeric characters, '-' or '.'"
                          />
                        </div>
                      </div>
                    </div>
                    <div className="confirmation-modal-buttons">
                      <CdsButton type="button" onClick={closeNewNSModal}>
                        Cancel
                      </CdsButton>
                      <CdsButton status="primary" type="submit">
                        Submit
                      </CdsButton>
                    </div>
                  </form>
                </div>
              </Modal>
              <CdsButton
                disabled={!canCreateNS}
                title={
                  canCreateNS
                    ? "Create a new namespace in the current cluster"
                    : "You don't have permission to create namespaces on the cluster"
                }
                status="inverse"
                size="sm"
                action="flat"
                className="flat-btn"
                onClick={openNewNSModal}
              >
                Create Namespace
              </CdsButton>
            </div>
          </div>
          <div className="dropdown-menu-padding">
            <CdsButton status="primary" size="sm" onClick={changeContext}>
              Change Context
            </CdsButton>
          </div>
        </div>
      </div>
    </>
  );
}

export default ContextSelector;
