import * as React from "react";
import { Link } from "react-router-dom";
import Card, { CardBlock, CardFooter, CardHeader } from "../js/Card";
import Column from "../js/Column";
import Row from "../js/Row";

import placeholder from "../../placeholder.png";
import "./InfoCard.v2.css";

export interface IInfoCardProps {
  title: string;
  info: string | JSX.Element;
  link?: string;
  icon?: string;
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
    subIcon,
  } = props;
  const icon = props.icon ? props.icon : placeholder;
  return (
    <Column span={[12, 6, 4, 3]}>
      <Card clickable={true}>
        <Link to={link || "#"}>
          <CardHeader>
            <>
              <div className="card-title">{title}</div>
              {subIcon && <img src={subIcon} alt="icon" />}
            </>
          </CardHeader>
          <CardBlock>
            <Row>
              <Column span={3}>
                <div className="card-icon">
                  <img src={icon} alt="icon" sizes="64px" />
                </div>
              </Column>
              <Column span={9}>
                <span>{description}</span>
              </Column>
            </Row>
          </CardBlock>
          <CardFooter>
            <Row>
              <div className="kubeapps-card-footer">
                <div className="kubeapps-card-footer-col1">{info}</div>
                <div className="kubeapps-card-footer-col2">
                  <div className="footer-tags">
                    {tag1Content && (
                      <span className={`label ${tag1Class || "label-info"}`}>{tag1Content}</span>
                    )}
                    {tag2Content && (
                      <span className={`label ${tag2Class || "label-info"}`}>{tag2Content}</span>
                    )}
                  </div>
                </div>
              </div>
            </Row>
          </CardFooter>
        </Link>
      </Card>
    </Column>
  );
}

export default InfoCard;
