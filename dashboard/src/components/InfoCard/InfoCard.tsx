import * as React from "react";

import { Link } from "react-router-dom";
import placeholder from "../../placeholder.png";
import Card, { CardContent, CardIcon } from "../Card";
import "./InfoCard.css";

export interface IServiceInstanceCardProps {
  title: string;
  info: string;
  link?: string;
  icon?: string;
  banner?: string;
  description?: string | JSX.Element;
  tag1Class?: string;
  tag1Content?: string | JSX.Element;
  tag2Class?: string;
  tag2Content?: string | JSX.Element;
}

const InfoCard: React.SFC<IServiceInstanceCardProps> = props => {
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
  } = props;
  const icon = props.icon ? props.icon : placeholder;
  return (
    <Card responsive={true} className="ListItem">
      <Link to={link || "#"} title={title} className="ListItem__header">
        <CardIcon icon={icon} />
        {banner && <p className="ListItem__banner">{banner}</p>}
      </Link>
      <CardContent>
        <div className="ListItem__content">
          <Link to={link || "#"} title={title}>
            <h3 className="ListItem__content__title type-big">{title}</h3>
          </Link>
          {description}
          <div className="ListItem__content__info">
            <p className="margin-reset type-small padding-t-tiny type-color-light-blue">{info}</p>
            <div>
              {tag1Content && (
                <span
                  className={`ListItem__content__info_tag ListItem__content__info_tag-1 type-small type-color-white padding-t-tiny padding-h-normal ${tag1Class ||
                    ""}`}
                >
                  {tag1Content}
                </span>
              )}
              {tag2Content && (
                <span
                  className={`ListItem__content__info_tag ListItem__content__info_tag-2 type-small type-color-white padding-t-tiny padding-h-normal margin-b-normal ${tag2Class ||
                    ""}`}
                >
                  {tag2Content}
                </span>
              )}
            </div>
          </div>
        </div>
      </CardContent>
      {props.children}
    </Card>
  );
};

export default InfoCard;
