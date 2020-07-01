import { ClarityIcons, infoCircleIcon } from "@clr/core/icon-shapes";
import * as React from "react";
import { Link } from "react-router-dom";
import { CdsIcon } from "../Clarity/clarity";

import placeholder from "../../placeholder.png";
import "./InfoCard.v2.css";

ClarityIcons.addIcons(infoCircleIcon);

export interface IInfoCardProps {
  title: string;
  info: string | JSX.Element;
  link?: string;
  icon?: string;
  banner?: string;
  subIcon?: string;
  description?: string | JSX.Element;
  tag1Class?: string;
  tag1Content?: string | JSX.Element;
  tag2Class?: string;
  tag2Content?: string | JSX.Element;
}

function InfoCard(props: IInfoCardProps) {
  const {
    title,
    link,
    info,
    description,
    tag1Content,
    tag1Class,
    tag2Content,
    tag2Class,
    banner,
    subIcon,
  } = props;
  const icon = props.icon ? props.icon : placeholder;
  return (
    <div className="clr-col-lg-3 clr-col-12">
      <Link to={link || "#"} className="card clickable">
        <div className="card-header">{title}</div>
        <div className="card-block">
          <div className="card-media-block">
            <img src={icon} className="card-media-image" alt="icon" />
            <div className="card-media-description">
              <span className="card-media-text">{description}</span>
            </div>
          </div>
          {banner && (
            <div className="alert alert-info alert-sm" role="alert">
              <div className="alert-items">
                <div className="alert-item static">
                  <div className="alert-icon-wrapper">
                    <CdsIcon className="alert-icon" shape="info-circle" />
                  </div>
                  <span className="alert-text">{banner}</span>
                </div>
              </div>
            </div>
          )}
        </div>
        <div className="card-footer">
          <div className="clr-row">
            <div className="clr-col-6" style={{ padding: 0 }}>
              {info}
            </div>
            <div className="clr-col-6 label-section">
              {tag1Content && (
                <span className={`label ${tag1Class || "label-info"}`}>{tag1Content}</span>
              )}
              {tag2Content && (
                <span className={`label ${tag2Class || "label-info"}`}>{tag2Content}</span>
              )}
              {subIcon && <img src={subIcon} alt="icon" />}
            </div>
          </div>
        </div>
      </Link>
    </div>
  );
}

export default InfoCard;
