import { CdsButton } from "@clr/react/button";
import { CdsFormGroup } from "@clr/react/forms";
import { CdsIcon } from "@clr/react/icon";
import { CdsInput } from "@clr/react/input";
import { CdsModal, CdsModalActions, CdsModalContent, CdsModalHeader } from "@clr/react/modal";
import actions from "actions";
import Alert from "components/js/Alert";
import Column from "components/js/Column";
import { push } from "connected-react-router";
import React, { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { definedNamespaces } from "shared/Namespace";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import useOutsideClick from "../js/hooks/useOutsideClick/useOutsideClick";
import Row from "../js/Row";
import "./ContextSelector.css";

function ContextSelector() {
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
    if (namespaceSelected && namespaceSelected !== definedNamespaces.all) {
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
    dispatch(push(app.apps.list(cluster, namespace)));
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
      dispatch(push(app.apps.list(cluster, newNS)));
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
                    {namespaceSelected === definedNamespaces.all
                      ? "All Namespaces"
                      : namespaceSelected}
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
                <option key="kubeapps-dropdown-namespace-_all" value={definedNamespaces.all}>
                  All Namespaces
                </option>
              </select>
            </div>
            <div className="kubeapps-create-new-ns">
              <CdsModal hidden={!newNSModalIsOpen} closable={true} onCloseChange={closeNewNSModal}>
                <CdsModalHeader>Create a New Namespace</CdsModalHeader>
                {error && <Alert theme="danger">An error occurred: {error.error.message}</Alert>}
                <form onSubmit={createNewNS}>
                  <CdsModalContent>
                    <CdsFormGroup>
                      <CdsInput>
                        <label>Name:</label>
                        <input type="text" required={true} onChange={onChangeNewNS} />
                      </CdsInput>
                    </CdsFormGroup>
                  </CdsModalContent>
                  <CdsModalActions>
                    <CdsButton type="submit">Submit</CdsButton>
                  </CdsModalActions>
                </form>
              </CdsModal>
              <CdsButton
                disabled={!canCreateNS}
                title={canCreateNS ? "" : "missing permissions"}
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
