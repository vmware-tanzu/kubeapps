// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import React from "react";

export interface ICardTextProps {
  children: React.ReactNode;
}

const CardText = ({ children }: ICardTextProps) => {
  return <div className="card-text">{children}</div>;
};

export default CardText;
