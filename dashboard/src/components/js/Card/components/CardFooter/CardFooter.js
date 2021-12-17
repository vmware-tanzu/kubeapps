import PropTypes from "prop-types";
import React from "react";

const CardFooter = ({ children }) => <footer className="card-footer">{children}</footer>;

CardFooter.propTypes = {
  children: PropTypes.oneOfType([
    PropTypes.arrayOf(PropTypes.element),
    PropTypes.element,
    PropTypes.string,
  ]).isRequired,
};

export default CardFooter;
