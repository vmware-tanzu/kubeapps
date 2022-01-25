// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import cs from "classnames";
import PropTypes from "prop-types";
import React from "react";
import { Link } from "react-router-dom";
import Spinner from "../Spinner";
// Custom Styles
import "./Button.scss";

// Constants
export const ButtonThemes = {
  danger: "danger",
  inverse: "inverse",
  primary: "primary",
  success: "success",
  warning: "warning",
};

export const ButtonTypes = {
  button: "button",
  submit: "submit",
  reset: "reset",
};

/**
 * @see https://clarity.design/documentation/buttons#examples
 */
const Button = ({
  block,
  className,
  children,
  disabled,
  icon,
  externalLink,
  flat,
  link,
  loading,
  loadingText,
  onClick,
  outline,
  rel,
  target,
  theme,
  title,
  type,
  small,
}) => {
  const css = cs("btn button", className, {
    [`btn-${theme}`]: !flat && !outline,
    [`btn-${theme}-outline`]: outline,
    [`btn-link`]: flat,
    [`btn-icon`]: icon,
    [`btn-sm`]: small,
    [`btn-block`]: block,
  });

  if (externalLink != null) {
    // External URL
    return (
      <a
        className={css}
        title={title}
        href={externalLink}
        disabled={disabled}
        rel={rel}
        target={target}
      >
        {children}
      </a>
    );
  } else if (link != null) {
    // Internal URL
    return (
      <Link className={css} title={title} to={link} disabled={disabled}>
        {children}
      </Link>
    );
  } else {
    // Normal Button
    return (
      <button className={css} title={title} onClick={onClick} disabled={disabled} type={type}>
        {loading ? <Spinner inline text={loadingText} /> : children}
      </button>
    );
  }
};

Button.propTypes = {
  block: PropTypes.bool,
  className: PropTypes.string,
  children: PropTypes.node.isRequired,
  disabled: PropTypes.bool,
  icon: PropTypes.bool,
  externalLink: PropTypes.string,
  flat: PropTypes.bool,
  link: PropTypes.string,
  loading: PropTypes.bool,
  loadingText: PropTypes.string,
  onClick: PropTypes.func,
  outline: PropTypes.bool,
  rel: PropTypes.string,
  small: PropTypes.bool,
  target: PropTypes.string,
  theme: PropTypes.oneOf(Object.values(ButtonThemes)).isRequired,
  title: PropTypes.string,
  type: PropTypes.oneOf(Object.values(ButtonTypes)).isRequired,
};

Button.defaultProps = {
  block: false,
  className: "",
  disabled: false,
  flat: false,
  icon: false,
  outline: false,
  loading: false,
  small: false,
  theme: ButtonThemes.primary,
  title: "",
  type: ButtonTypes.button,
};

export default Button;
