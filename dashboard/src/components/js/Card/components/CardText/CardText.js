import PropTypes from "prop-types";
import React from "react";

const CardText = ({ children }) => <div className="card-text">{children}</div>;

CardText.propTypes = {
  children: PropTypes.node.isRequired,
};

export default CardText;
