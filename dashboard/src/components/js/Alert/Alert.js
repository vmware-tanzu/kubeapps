// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import cs from "classnames";
import PropTypes from "prop-types";
import React from "react";
import Button from "../Button";
import "./Alert.scss";
import AlerTypes from "./Alert.types";

// Constants
export const AlertThemes = {
  danger: "danger",
  info: "info",
  success: "success",
  warning: "warning",
};

export const AlertIcons = {
  danger: "exclamation-circle",
  info: "info-circle",
  success: "check-circle",
  warning: "exclamation-triangle",
};

/**
 * https://v2.clarity.design/alerts
 */
const Alert = ({ app, action, children, customIcon, compact, theme, onClick, onClose }) => {
  // We only want to announce the section on app level alerts
  const accessibilityProps = app
    ? {
        "aria-label": "Please, read the following important information",
      }
    : {};
  const css = cs(`alert alert-${theme}`, {
    "alert-app-level": app,
    "alert-sm": compact,
  });
  // Custom icons are only available for app level
  let icon = app && customIcon != null ? customIcon : AlertIcons[theme];

  // If we provide an action, we should use alertdialog as the role
  // https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Roles/Alert_Role
  if (onClose != null || onClick != null) {
    accessibilityProps.role = "alertdialog";
  } else {
    accessibilityProps.role = "alert";
  }

  return (
    <section className={css} {...accessibilityProps}>
      <div className="alert-items">
        <div className="alert-item static">
          <div className="alert-icon-wrapper">
            <CdsIcon class="alert-icon" shape={icon} aria-hidden="true"></CdsIcon>
          </div>
          <div className="alert-text">{children}</div>
          {onClick != null && action != null && (
            <div className="alert-actions">
              <Button className="alert-action" onClick={onClick}>
                {action}
              </Button>
            </div>
          )}
        </div>
      </div>
      {onClose != null && (
        <button type="button" onClick={onClose} className="close" aria-label="Close alert">
          <CdsIcon aria-hidden="true" shape="close"></CdsIcon>
        </button>
      )}
    </section>
  );
};

Alert.propTypes = {
  action: PropTypes.string,
  app: PropTypes.bool,
  children: React.ReactNode || PropTypes.node.isRequired,
  compact: PropTypes.bool,
  customIcon: AlerTypes.customIcon,
  theme: AlerTypes.theme,
  onClick: PropTypes.func,
  onClose: PropTypes.func,
};

Alert.defaultProps = {
  app: false,
  compact: false,
  theme: AlertThemes.info,
};

export default Alert;
