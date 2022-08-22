// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import { CdsToggle } from "@cds/react/toggle";
import actions from "actions";
import { getThemeFile } from "components/HeadManager/HeadManager";
import { useEffect, useRef, useState } from "react";
import { Helmet } from "react-helmet";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router-dom";
import { CSSTransition } from "react-transition-group";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { SupportedThemes } from "shared/Config";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import operatorIcon from "icons/olm-icon-white.svg";
import { IClustersState } from "../../reducers/cluster";
import useOutsideClick from "../js/hooks/useOutsideClick/useOutsideClick";
import Row from "../js/Row";
import "./Menu.css";

export interface IContextSelectorProps {
  clusters: IClustersState;
  appVersion: string;
  logout: () => void;
}

function Menu({ clusters, appVersion, logout }: IContextSelectorProps) {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const [open, setOpen] = useState(false);
  const currentCluster = clusters.clusters[clusters.currentCluster];
  const namespaceSelected = currentCluster.currentNamespace;
  // Control when users click outside
  const ref = useRef(null);
  useOutsideClick(setOpen, [ref], open);

  const {
    config: { theme, featureFlags },
  } = useSelector((state: IStoreState) => state);

  const toggleOpen = () => setOpen(!open);

  const toggleTheme = () => {
    const newTheme = theme === SupportedThemes.dark ? SupportedThemes.light : SupportedThemes.dark;
    dispatch(actions.config.setUserTheme(newTheme));
  };

  useEffect(() => {
    document.body.setAttribute("cds-theme", theme);
  }, [theme]);

  /* eslint-disable jsx-a11y/label-has-associated-control */
  return (
    <>
      <Helmet>
        {/*  Override the clarity-ui css style */}
        <link rel="stylesheet" type="text/css" href={getThemeFile(SupportedThemes[theme])} />
      </Helmet>

      <div className={open ? "drawer-backdrop" : ""} />
      <div className={`dropdown kubeapps-menu ${open ? "open" : ""}`} ref={ref}>
        <button
          className="kubeapps-nav-link"
          onClick={toggleOpen}
          aria-expanded={open}
          aria-haspopup="menu"
        >
          <Row>{<CdsIcon size="md" shape="applications" solid={true} />}</Row>
        </button>
        <CSSTransition in={open} timeout={200} classNames="transition-drawer">
          <div className="dropdown-menu dropdown-configuration-menu" role="menu" hidden={!open}>
            <div>
              <label className="dropdown-menu-padding dropdown-menu-label">Administration</label>
              <Link
                to={app.config.pkgrepositories(clusters.currentCluster, namespaceSelected)}
                className="dropdown-menu-link"
                onClick={toggleOpen}
              >
                <div className="dropdown-menu-item" role="menuitem">
                  <CdsIcon solid={true} size="md" shape="library" />{" "}
                  <span>Package Repositories</span>
                </div>
              </Link>
              <div className="dropdown-divider" role="separator" />
              {featureFlags?.operators && (
                <Link
                  to={app.config.operators(clusters.currentCluster, namespaceSelected)}
                  className="dropdown-menu-link"
                  onClick={toggleOpen}
                >
                  <div className="dropdown-menu-item" role="menuitem">
                    <img src={operatorIcon} alt="operators-icon" />
                    <span>Operators</span>
                  </div>
                </Link>
              )}
              {featureFlags?.operators && <div className="dropdown-divider" role="separator" />}
            </div>
            <div>
              <div className="dropdown-menu-subtext">
                Made with <CdsIcon size="sm" shape="heart" solid={true} /> by VMware and{" "}
                <a
                  href="https://github.com/vmware-tanzu/kubeapps/graphs/contributors"
                  className="type-color-white"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  contributors
                </a>
                .
                <br />
                {appVersion}
                <br />
                <Link to={"/docs"}>
                  API documentation portal <CdsIcon size="sm" shape="network-globe" solid={true} />
                </Link>
                <CdsToggle className="dropdown-theme-toggle" control-align="right">
                  <label>
                    <span className="toggle-label-text">
                      <CdsIcon
                        size="sm"
                        shape={theme === SupportedThemes.dark ? "moon" : "sun"}
                        solid={true}
                      />
                    </span>
                  </label>
                  <input
                    type="checkbox"
                    onChange={toggleTheme}
                    checked={theme === SupportedThemes.dark}
                  />
                </CdsToggle>
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
