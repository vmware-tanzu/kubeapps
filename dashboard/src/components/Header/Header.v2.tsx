import * as React from "react";
import { NavLink } from "react-router-dom";

import logo from "../../logo.svg";
import { IClustersState } from "../../reducers/cluster";
import { app } from "../../shared/url";
import ContextSelector from "./ContextSelector";
import "./Header.v2.css";
import Menu from "./Menu";

interface IHeaderProps {
  authenticated: boolean;
  fetchNamespaces: (cluster: string) => void;
  logout: () => void;
  clusters: IClustersState;
  defaultNamespace: string;
  appVersion: string;
  push: (path: string) => void;
  setNamespace: (ns: string) => void;
  createNamespace: (cluster: string, ns: string) => Promise<boolean>;
  getNamespace: (cluster: string, ns: string) => void;
}

function Header(props: IHeaderProps) {
  const {
    appVersion,
    clusters,
    authenticated: showNav,
    defaultNamespace,
    fetchNamespaces,
    createNamespace,
    getNamespace,
    setNamespace,
    logout,
  } = props;
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
              <ContextSelector
                clusters={clusters}
                fetchNamespaces={fetchNamespaces}
                getNamespace={getNamespace}
                createNamespace={createNamespace}
                defaultNamespace={defaultNamespace}
                setNamespace={setNamespace}
              />
              <Menu
                clusters={clusters}
                defaultNamespace={defaultNamespace}
                appVersion={appVersion}
                logout={logout}
              />
            </section>
          )}
        </header>
      </div>
    </section>
  );
}

export default Header;
