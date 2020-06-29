import {
  angleIcon,
  applicationsIcon,
  ClarityIcons,
  clusterIcon,
  fileGroupIcon,
} from "@clr/core/icon-shapes";
import * as React from "react";
import { NavLink } from "react-router-dom";
import { CdsIcon } from "../Clarity/clarity";

import logo from "../../logo.svg";
import { IClusterState } from "../../reducers/cluster";
import { app } from "../../shared/url";
import "./Header.v2.css";

ClarityIcons.addIcons(applicationsIcon, clusterIcon, fileGroupIcon, angleIcon);

interface IHeaderProps {
  authenticated: boolean;
  fetchNamespaces: () => void;
  logout: () => void;
  cluster: IClusterState;
  defaultNamespace: string;
  push: (path: string) => void;
  setNamespace: (ns: string) => void;
  createNamespace: (ns: string) => Promise<boolean>;
  getNamespace: (ns: string) => void;
}

class Header extends React.Component<IHeaderProps> {
  public render() {
    const { cluster, authenticated: showNav } = this.props;

    const routesToRender = [
      { title: "Applications", path: app.apps.list(cluster.currentNamespace), external: false },
      { title: "Catalog", path: app.catalog(cluster.currentNamespace), external: false },
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
                <div className="dropdown bottom-right kubeapps-align-center kubeapps-nav-link kubeapps-dropdown">
                  <div className="clr-row">
                    <div className="clr-col-10">
                      <span>Current Context</span>
                      <div>
                        {/* TODO(andresmgot): Add cluster */}
                        <CdsIcon size="sm" shape="cluster" inverse={true} />
                        <span className="kubeapps-dropdown-text">default</span>
                        <CdsIcon size="sm" shape="file-group" inverse={true} />
                        <span className="kubeapps-dropdown-text">{cluster.currentNamespace}</span>
                      </div>
                    </div>
                    <div className="clr-col-2 kubeapps-align-center">
                      <CdsIcon shape="angle" inverse={true} />
                    </div>
                  </div>
                </div>
                <button
                  className="kubeapps-nav-link nav-icon"
                  aria-label="VMware Cloud services configuration"
                >
                  <CdsIcon size="lg" shape="applications" solid={true} />
                </button>
              </section>
            )}
          </header>
        </div>
      </section>
    );
  }
}

export default Header;
