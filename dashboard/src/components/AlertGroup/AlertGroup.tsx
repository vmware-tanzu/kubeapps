// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsAlert, CdsAlertActions, CdsAlertGroup } from "@cds/react/alert";
import React from "react";
import "./AlertGroup.css";

export interface IAlertGroupProps {
  children: React.ReactNode;
  alertActions?: string | JSX.Element;
  closable?: boolean;
  status?: "neutral" | "info" | "success" | "warning" | "danger" | "alt" | "loading";
  type?: "default" | "banner" | "light";
  size?: "default" | "sm";
  withMargin?: boolean;
}

// Opinionated wrapper of Clarity AlertGroup
// https://storybook.core.clarity.design/?path=/story/components-alert--page
export default function AlertGroup({
  alertActions,
  closable = true,
  status = "danger",
  type = "default",
  size = "default",
  withMargin = true,
  children,
}: IAlertGroupProps) {
  const [closed, setClosed] = React.useState(false);
  const close = () => setClosed(true);
  return (
    <CdsAlertGroup
      className={withMargin ? "alert-group-margin" : ""}
      status={status}
      type={type}
      hidden={closed}
      size={size}
    >
      <CdsAlert closable={closable} onCloseChange={close}>
        {children}
        <CdsAlertActions>{alertActions}</CdsAlertActions>
      </CdsAlert>
    </CdsAlertGroup>
  );
}
