// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import React from "react";

export interface ICardFooterProps {
  children: React.ReactNode;
}

const CardFooter = ({ children }: ICardFooterProps) => {
  return <footer className="card-footer">{children}</footer>;
};

export default CardFooter;
