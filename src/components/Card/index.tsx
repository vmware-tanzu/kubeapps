import * as React from "react";
import { Link } from "react-router-dom";

import "./Card.css";

export interface ICardProps {
  header: string | JSX.Element | JSX.Element[];
  body: string | JSX.Element | JSX.Element[];
  button?: JSX.Element;
  buttonText?: string | JSX.Element;
  onClick?: () => (...args: any[]) => Promise<any>;
  linkTo?: string;
  notes?: string | JSX.Element;
  icon?: string;
}

export const CardContainer = (props: any) => {
  return <div className="CardContainer">{props.children}</div>;
};

export const Card = (props: ICardProps) => {
  const { header, body, buttonText, onClick, linkTo, notes, icon } = props;
  let button = props.button ? (
    props.button
  ) : (
    <button onClick={onClick} className="button button-primary" style={{ width: "fit-content" }}>
      {buttonText}
    </button>
  );
  if (linkTo) {
    button = <Link to={linkTo}>{button}</Link>;
  }
  return (
    <div className="Card">
      <h5 className="Card__header">{header}</h5>
      <div className="Card__body">{body}</div>
      <div className="Card__notes">{notes}</div>
      <div className="Card__button">{button}</div>
      <div className="Card__icon">
        <img className="Card__img" src={icon} />
      </div>
    </div>
  );
};
