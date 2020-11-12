import * as React from "react";
import { NavLink } from "react-router-dom";

import logo from "../../logo.svg";
import { IClustersState } from "../../reducers/cluster";
import { app } from "../../shared/url";
import ContextSelector from "./ContextSelector";
import "./Header.css";
import Menu from "./Menu";

interface IHeaderProps {
  authenticated: boolean;
  logout: () => void;
  clusters: IClustersState;
  appVersion: string;
  push: (path: string) => void;
}

function Header(props: IHeaderProps) {
  const { appVersion, clusters, authenticated: showNav, logout } = props;
  const cluster = clusters.clusters[clusters.currentCluster];

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
          <div className="branding">
            <NavLink to="/">
              <img src={logo} alt="Kubeapps logo" className="kubeapps__logo" />
            </NavLink>
          </div>
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
