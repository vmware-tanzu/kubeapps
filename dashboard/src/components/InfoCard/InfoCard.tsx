// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Icon from "components/Icon/Icon";
import { Link } from "react-router-dom";
import placeholder from "icons/placeholder.svg";
import Card, { CardBlock, CardFooter, CardHeader } from "../js/Card";
import Column from "../js/Column";
import Row from "../js/Row";
import "./InfoCard.css";

export interface IInfoCardProps {
  title: string;
  info: string | JSX.Element;
  link?: string;
  icon?: string;
  bgIcon?: string;
  description?: string | JSX.Element;
  tooltip?: JSX.Element;
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
    tooltip,
    tag1Content,
    tag1Class,
    tag2Content,
    tag2Class,
    bgIcon,
  } = props;
  const icon = props.icon ? props.icon : placeholder;
  return (
    <Column span={[12, 6, 4, 3]}>
      <Card clickable={true}>
        <Link to={link || "#"}>
          <CardHeader>
            <div className="info-card-header">
              <div className="card-title">{title}</div>
              {tooltip ? <div className="card-tooltip">{tooltip}</div> : <></>}
            </div>
          </CardHeader>
          <CardBlock>
            <div className="info-card-block">
              <div className="card-icon">
                <Icon icon={icon} />
              </div>
              <div className="card-description-wrapper">
                <div className="card-description">
                  <span>{description}</span>
                </div>
              </div>
              {bgIcon ? (
                <div className="bg-img">
                  <img src={bgIcon} alt="bg-img" />
                </div>
              ) : (
                <></>
              )}
            </div>
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
