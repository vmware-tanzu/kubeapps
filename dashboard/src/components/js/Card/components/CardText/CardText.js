import React from "react";
import PropTypes from "prop-types";

const CardText = ({ children }) => <div className="card-text">{children}</div>;

CardText.propTypes = {
  children: PropTypes.node.isRequired,
};

export default CardText;
