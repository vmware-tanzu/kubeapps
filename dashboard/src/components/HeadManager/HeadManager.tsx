import React from "react";
import { Helmet } from "react-helmet";
import { useIntl } from "react-intl";

interface IHeadManagerProps {
  theme: SupportedThemes;
  children: React.ReactNode;
}

export enum SupportedThemes {
  dark = "dark",
  light = "light",
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
export default function HeadManager({ theme, children }: IHeadManagerProps) {
  const intl = useIntl();

  document.body.setAttribute("cds-theme", theme); // sets the initial cds theme
  localStorage.setItem("theme", theme); // persist the initial theme decision

  return (
    <>
      <Helmet>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />

        <link rel="icon" type="image/png" href="./favicon-196x196.png" sizes="196x196" />
        <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
        <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
        <link rel="icon" type="image/png" href="./favicon-128.png" sizes="128x128" />
        <link rel="apple-touch-icon" href="./favicon-196x196.png" />
        <link rel="manifest" href="./manifest.json" />

        {/*  Allow to load custom styling different. The dashboard webserver will return this style file.  */}
        <link rel="stylesheet" type="text/css" href="./custom_style.css" />

        {/*  Set the clarity-ui css style */}
        <link rel="stylesheet" type="text/css" href={getThemeFile(theme)} />

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
