// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { HTMLAttributes } from "react";

export interface IRowProps {
  isList?: boolean;
  children: React.ReactNode;
}

export function Row({ isList, children }: IRowProps) {
  const innerProps: HTMLAttributes<HTMLDivElement> = {
    className: "clr-row",
    role: isList ? "list" : "",
  };

  return <div {...innerProps}>{children}</div>;
}
