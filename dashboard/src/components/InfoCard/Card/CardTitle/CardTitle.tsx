// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import React from "react";
import "./CardTitle.scss";

export interface ICardTitleProps {
  children: React.ReactNode;
  level: 1 | 2 | 3 | 4 | 5 | 6;
}

const CardTitle = ({ level, children }: ICardTitleProps) => {
  return React.createElement(`h${level}`, { className: "card-title" }, children);
};

export default CardTitle;
