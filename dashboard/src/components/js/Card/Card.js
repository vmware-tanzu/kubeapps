// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import cs from "classnames";
import PropTypes from "prop-types";
import React from "react";

const Card = ({ children, clickable, htmlTag, onClick }) => {
  const cssClass = cs("card", {
    clickable: clickable || typeof onClick === "function",
  });

  // Common props for the element
  const innerProps = {
    className: cssClass,
    onClick,
  };

  // Make the card focuseable by keyboard navigation. I'm not adding it based on the
  // clickable prop because I assume there's something around the card that manages
  // the onClick event.
  if (typeof onClick === "function") {
    innerProps["tabIndex"] = "0";

    // Runs onClick when the user types `enter`
    innerProps["onKeyPress"] = e => {
      // Enter
      if (e.key === "Enter") {
        onClick(e);
      }
    };

    // Runs onClick when the user types `space`
    innerProps["onKeyUp"] = e => {
      // Space
      if (e.key === " ") {
        onClick(e);
      }
    };
  }

  return React.createElement(htmlTag, innerProps, children);
};

Card.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.element), PropTypes.element])
    .isRequired,
  clickable: PropTypes.bool,
  htmlTag: PropTypes.string,
  onClick: PropTypes.func,
};

Card.defaultProps = {
  clickable: false,
  htmlTag: "div",
};

export default Card;
