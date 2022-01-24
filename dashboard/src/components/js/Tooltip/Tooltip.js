// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import cs from "classnames";
import PropTypes from "prop-types";
import { useState } from "react";
import "./Tooltip.scss";

// Constants
export const TooltipPositions = {
  topRight: "top-right",
  topLeft: "top-left",
  bottomRight: "bottom-right",
  bottomLeft: "bottom-left",
  right: "right",
  left: "left",
};

const Tooltip = ({ children, position, icon, iconProps, id, label, extraSmall, small, large }) => {
  const [open, setOpen] = useState(false);
  let timer;
  const openTooltip = () => {
    setOpen(true);
    clearTimeout(timer);
  };

  const closeTooltip = () => {
    timer = setTimeout(() => {
      setOpen(false);
    }, 1000);
  };

  const escapeFromTooltip = e => {
    if (e.key === "Escape") setOpen(false);
  };

  const css = cs("tooltip", `tooltip-${position}`, {
    "tooltip-open": open,
    "tooltip-xs": extraSmall,
    "tooltip-md": !extraSmall && !small && !large,
    "tooltip-sm": small,
    "tooltip-lg": large,
  });

  // the 'tooltip' role does have the inherited props 'aria-haspopup' and 'aria-expanded'
  /* eslint-disable jsx-a11y/role-supports-aria-props */
  /* eslint-disable jsx-a11y/no-interactive-element-to-noninteractive-role */
  return (
    <button
      className={css}
      role="tooltip"
      aria-haspopup="true"
      aria-expanded={open}
      aria-label={label}
      aria-describedby={id}
      onMouseEnter={openTooltip}
      onMouseLeave={closeTooltip}
      onKeyUp={escapeFromTooltip}
      onFocus={openTooltip}
      onBlur={closeTooltip}
    >
      <CdsIcon size="24" role="none" shape={icon} {...iconProps}></CdsIcon>
      <span id={id} className="tooltip-content" aria-hidden={!open}>
        {children}
      </span>
    </button>
  );
};

Tooltip.propTypes = {
  children: PropTypes.node,
  position: PropTypes.string,
  icon: PropTypes.string,
  iconProps: PropTypes.object,
  // Uniq ID for the tooltip in the view where it's used
  id: PropTypes.string.isRequired,
  // Label to announce the tooltip. It's used for accessibility
  label: PropTypes.string.isRequired,
  extraSmall: PropTypes.bool,
  small: PropTypes.bool,
  large: PropTypes.bool,
};

Tooltip.defaultProps = {
  children: PropTypes.node,
  position: TooltipPositions.topRight,
  icon: "info-circle",
  iconProps: {},
  extraSmall: false,
  small: false,
  large: false,
};

export default Tooltip;
