// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import React from "react";

export interface ICardBlockProps {
  children: React.ReactNode;
}

const CardBlock = ({ children }: ICardBlockProps) => {
  return <div className="card-block">{children}</div>;
};

export default CardBlock;
