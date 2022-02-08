// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { NavLink } from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import ContextSelector from "./ContextSelector";
import "./Header.css";
import Menu from "./Menu";

function Header() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const {
    auth: { authenticated },
    clusters,
    config: { appVersion },
  } = useSelector((state: IStoreState) => state);
  const cluster = clusters.clusters[clusters.currentCluster];
  const showNav = authenticated && clusters.currentCluster && cluster.currentNamespace;
  const logout = () => dispatch(actions.auth.logout());

  useEffect(() => {
    if (authenticated && clusters.currentCluster) {
      dispatch(actions.namespace.fetchNamespaces(clusters.currentCluster));
      dispatch(actions.namespace.canCreate(clusters.currentCluster));
    }
  }, [dispatch, authenticated, clusters.currentCluster]);

  const routesToRender = [
    {
      title: "Applications",
      path: app.apps.list(clusters.currentCluster, cluster.currentNamespace),
      external: false,
    },
    {
      title: "Catalog",
      path: app.catalog(clusters.currentCluster, cluster.currentNamespace),
      external: false,
    },
  ];
  return (
    <section>
      <div className="container">
        <header className="header header-7">
          <NavLink to="/">
            <div className="kubeapps-logo">
              <span className="sr-only">Homepage</span>
            </div>
          </NavLink>
          {showNav && (
            <nav className="header-nav">
              {routesToRender.map(route => {
                const { path, title } = route;
                return (
                  <NavLink
                    key={path}
                    to={path}
                    activeClassName="active"
                    className="nav-link nav-text"
                  >
                    {title}
                  </NavLink>
                );
              })}
            </nav>
          )}
          {showNav && (
            <section className="header-actions">
              <ContextSelector />
              <Menu clusters={clusters} appVersion={appVersion} logout={logout} />
            </section>
          )}
        </header>
      </div>
    </section>
  );
}

export default Header;
