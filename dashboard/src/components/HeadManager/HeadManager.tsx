import React from "react";
import { Helmet } from "react-helmet";

interface IHeadManagerProps {
  children: React.ReactNode;
}

export default function HeadManager({ children }: IHeadManagerProps) {
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

        <meta name="theme-color" content="#304250" />
        <meta
          name="Kubeapps"
          content="A web-based UI for deploying and managing applications in Kubernetes clusters"
        />
        <title>Kubeapps Dashboard</title>
      </Helmet>
      {children}
    </>
  );
}
