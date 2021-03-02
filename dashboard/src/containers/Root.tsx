import Header from "components/Header";
import HeadManager, { SupportedThemes } from "components/HeadManager/HeadManager";
import Layout from "components/Layout";
import LoadingWrapper from "components/LoadingWrapper";
import { ConnectedRouter } from "connected-react-router";
import React, { Suspense, useEffect, useState } from "react";
import { IntlProvider } from "react-intl";
import { Provider } from "react-redux";
import I18n, { ISupportedLangs } from "shared/I18n";
import store, { history } from "../store";
// TODO(andresmgot): Containers should be no longer needed, replace them when possible
import ConfigLoaderContainer from "./ConfigLoaderContainer";
import Routes from "./RoutesContainer";

async function initLocale() {
  const fullLang = (navigator.languages && navigator.languages[0]) || navigator.language;
  const lang = fullLang.toLowerCase().split(/[_-]+/)[0];
  return await I18n.getCustomConfig(ISupportedLangs[lang]);
}

function getInitialTheme() {
  const userPreferredTheme = localStorage.getItem("theme");
  const browserPreferredTheme =
    window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches
      ? SupportedThemes.dark
      : SupportedThemes.light;
  return (userPreferredTheme as SupportedThemes) || browserPreferredTheme || SupportedThemes.light;
}

function Root() {
  const [i18nConfig, setI18nConfig] = useState(I18n.getDefaultConfig());
  const theme = getInitialTheme();

  useEffect(() => {
    initLocale().then(customI18nConfig => setI18nConfig(customI18nConfig));
  }, []);

  return (
    <Provider store={store}>
      <ConfigLoaderContainer>
        <ConnectedRouter history={history}>
          <Suspense fallback={LoadingWrapper}>
            <IntlProvider
              locale={i18nConfig.locale}
              key={i18nConfig.locale}
              messages={i18nConfig.messages}
              defaultLocale="en"
            >
              <HeadManager theme={theme}>
                <Layout headerComponent={Header}>
                  <Routes />
                </Layout>
              </HeadManager>
            </IntlProvider>
          </Suspense>
        </ConnectedRouter>
      </ConfigLoaderContainer>
    </Provider>
  );
}

export default Root;
