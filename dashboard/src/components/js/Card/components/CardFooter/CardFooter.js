import React from "react";
import PropTypes from "prop-types";

const CardFooter = ({ children }) => <footer className="card-footer">{children}</footer>;

CardFooter.propTypes = {
  children: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.element),
    PropTypes.element,
    PropTypes.string,
  ]).isRequired,
};

export default CardFooter;
