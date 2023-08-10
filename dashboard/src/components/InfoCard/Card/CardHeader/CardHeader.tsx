// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import "./CardHeader.scss";

export interface ICardHeaderProps {
  children: React.ReactNode;
  noBorder?: boolean;
}

const CardHeader = ({ children, noBorder = false }: ICardHeaderProps) => {
  return <header className={`card-header ${noBorder ? "no-border" : ""}`}>{children}</header>;
};

export default CardHeader;
