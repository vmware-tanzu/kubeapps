// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import React from "react";
import { Helmet } from "react-helmet";
import { useIntl } from "react-intl";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { SupportedThemes } from "shared/Config";
import { IStoreState } from "shared/types";

interface IHeadManagerProps {
  children: React.ReactNode;
}

export function getThemeFile(theme: SupportedThemes) {
  const lightThemeFile = "./clr-ui.min.css";
  const darkThemeFile = "./clr-ui-dark.min.css";
  switch (theme) {
    case SupportedThemes.light:
      return lightThemeFile;
    case SupportedThemes.dark:
      return darkThemeFile;
    default:
      return lightThemeFile;
  }
}
export default function HeadManager({ children }: IHeadManagerProps) {
  const intl = useIntl();
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();

  const {
    config: { theme },
  } = useSelector((state: IStoreState) => state);

  React.useEffect(() => {
    dispatch(actions.config.getTheme());
  }, [dispatch]);

  return (
    <>
      <Helmet>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />

        {/* generated with https://realfavicongenerator.net/ */}
        <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
        <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
        <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
        <link rel="manifest" href="/site.webmanifest" />
        <link rel="mask-icon" href="/safari-pinned-tab.svg" color="#0091da" />
        <link rel="shortcut icon" href="/favicon.ico" />
        <meta name="msapplication-TileColor" content="#0091da" />
        <meta name="msapplication-config" content="/browserconfig.xml" />
        <meta name="theme-color" content="#ffffff" />

        {/*  Allow to load custom styling different. The dashboard webserver will return this style file.  */}
        <link rel="stylesheet" type="text/css" href="./custom_style.css" />

        {/*  Set the clarity-ui css style */}
        <link rel="stylesheet" type="text/css" href={getThemeFile(SupportedThemes[theme])} />

        <meta name="theme-color" content="#304250" />
        <meta
          name={intl.formatMessage({ id: "Kubeapps", defaultMessage: "Kubeapps" })}
          content="A web-based UI for deploying and managing applications in Kubernetes clusters"
        />
        <title>
          {intl.formatMessage({ id: "Kubeapps", defaultMessage: "Kubeapps" })} Dashboard
        </title>
      </Helmet>
      {children}
    </>
  );
}
