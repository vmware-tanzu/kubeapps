import { push } from "connected-react-router";
import React, { useEffect, useRef, useState } from "react";
import { useDispatch } from "react-redux";

import { IClustersState } from "../../reducers/cluster";
import { CdsButton, CdsIcon } from "../Clarity/clarity";
import useOutsideClick from "../js/hooks/useOutsideClick/useOutsideClick";

import Column from "components/js/Column";
import { definedNamespaces } from "shared/Namespace";
import { app } from "shared/url";
import Row from "../js/Row";
import "./ContextSelector.css";

export interface IContextSelectorProps {
  clusters: IClustersState;
  defaultNamespace: string;
  fetchNamespaces: (cluster: string) => void;
  createNamespace: (cluster: string, ns: string) => Promise<boolean>;
  getNamespace: (cluster: string, ns: string) => void;
  setNamespace: (ns: string) => void;
}

function ContextSelector({
  clusters,
  defaultNamespace,
  fetchNamespaces,
  createNamespace,
  getNamespace,
  setNamespace,
}: IContextSelectorProps) {
  const dispatch = useDispatch();
  const [open, setOpen] = useState(false);
  const currentCluster = clusters.clusters[clusters.currentCluster];
  const namespaceSelected = currentCluster.currentNamespace || defaultNamespace;
  const [cluster, setStateCluster] = useState(clusters.currentCluster);
  const [namespace, setStateNamespace] = useState(namespaceSelected);
  // Control when users click outside
  const ref = useRef(null);
  useOutsideClick(setOpen, [ref], open);

  useEffect(() => {
    fetchNamespaces(clusters.currentCluster);
    if (namespaceSelected !== definedNamespaces.all) {
      getNamespace(clusters.currentCluster, namespaceSelected);
    }
  }, [fetchNamespaces, namespaceSelected, getNamespace, clusters.currentCluster]);

  const toggleOpen = () => setOpen(!open);
  const selectCluster = (event: React.ChangeEvent<HTMLSelectElement>) =>
    setStateCluster(event.target.value);
  const selectNamespace = (event: React.ChangeEvent<HTMLSelectElement>) =>
    setStateNamespace(event.target.value);
  const changeContext = () => {
    setNamespace(namespace);
    dispatch(push(app.apps.list(cluster, namespace)));
    setOpen(false);
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
        <div className="dropdown-menu" role="menu">
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
