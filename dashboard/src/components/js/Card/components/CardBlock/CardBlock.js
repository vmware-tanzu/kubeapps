import PropTypes from "prop-types";
import React from "react";

const CardBlock = ({ children }) => <div className="card-block">{children}</div>;

CardBlock.propTypes = {
  children: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.element),
    PropTypes.element,
    PropTypes.string,
  ]).isRequired,
};

export default CardBlock;
