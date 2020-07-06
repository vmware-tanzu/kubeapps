import React from "react";
import PropTypes from "prop-types";

const CardBlock = ({ children }) => <div className="card-block">{children}</div>;

CardBlock.propTypes = {
  children: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.element),
    PropTypes.element,
    PropTypes.string,
  ]).isRequired,
};

export default CardBlock;
