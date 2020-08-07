import React, { useRef, useState } from "react";
import { Link } from "react-router-dom";
import { CSSTransition } from "react-transition-group";

import { CdsButton, CdsIcon } from "../Clarity/clarity";
import useOutsideClick from "../js/hooks/useOutsideClick/useOutsideClick";

import { IClustersState } from "../../reducers/cluster";
import Row from "../js/Row";

import { app } from "shared/url";
import helmIcon from "../../icons/helm-white.svg";
import operatorIcon from "../../icons/operator-framework-white.svg";
import "./Menu.css";

export interface IContextSelectorProps {
  clusters: IClustersState;
  defaultNamespace: string;
  appVersion: string;
  logout: () => void;
}

function Menu({ clusters, defaultNamespace, appVersion, logout }: IContextSelectorProps) {
  const [open, setOpen] = useState(false);
  const currentCluster = clusters.clusters[clusters.currentCluster];
  const namespaceSelected = currentCluster.currentNamespace || defaultNamespace;
  // Control when users click outside
  const ref = useRef(null);
  useOutsideClick(setOpen, [ref], open);

  const toggleOpen = () => setOpen(!open);

  return (
    <>
      <div className={open ? "drawer-backdrop" : ""} />
      <div className={`dropdown kubeapps-menu ${open ? "open" : ""}`} ref={ref}>
        <button
          className="kubeapps-nav-link"
          onClick={toggleOpen}
          aria-expanded={open}
          aria-haspopup="menu"
        >
          <Row>
            <CdsIcon size="md" shape="applications" solid={true} inverse={true} />
          </Row>
        </button>
        <CSSTransition in={open} timeout={200} classNames="transition-drawer">
          <div className="dropdown-menu dropdown-configuration-menu" role="menu">
            <div>
              <label className="dropdown-menu-padding dropdown-menu-label">Administration</label>
              <Link
                to={app.config.apprepositories(clusters.currentCluster, namespaceSelected)}
                className="dropdown-menu-link"
                onClick={toggleOpen}
              >
                <div className="dropdown-menu-item" role="menuitem">
                  <img src={helmIcon} alt="helm-icon" />
                  <span>App Repositories</span>
                </div>
              </Link>
              <div className="dropdown-divider" role="separator" />
              <Link
                to={app.config.operators(clusters.currentCluster, namespaceSelected)}
                className="dropdown-menu-link"
                onClick={toggleOpen}
              >
                <div className="dropdown-menu-item" role="menuitem">
                  <img src={operatorIcon} alt="helm-icon" />
                  <span>Operators</span>
                </div>
              </Link>
              <div className="dropdown-divider" role="separator" />
            </div>
            <div>
              <div className="dropdown-menu-subtext">
                Made with <CdsIcon size="sm" shape="heart" inverse={true} solid={true} /> by Bitnami
                and{" "}
                <a
                  href="https://github.com/kubeapps/kubeapps/graphs/contributors"
                  className="type-color-white"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  contributors
                </a>
                .
                <br />
                {appVersion}
              </div>
              <div className="dropdown-menu-padding logout-button">
                <CdsButton status="primary" size="sm" action="outline" onClick={logout}>
                  Log out
                </CdsButton>
              </div>
            </div>
          </div>
        </CSSTransition>
      </div>
    </>
  );
}

export default Menu;
