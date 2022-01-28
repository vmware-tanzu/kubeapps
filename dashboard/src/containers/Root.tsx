// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Header from "components/Header";
import HeadManager from "components/HeadManager/HeadManager";
import Layout from "components/Layout";
import { ConnectedRouter } from "connected-react-router";
import { Suspense, useEffect, useState } from "react";
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

function Root() {
  const [i18nConfig, setI18nConfig] = useState(I18n.getDefaultConfig());

  useEffect(() => {
    initLocale().then(customI18nConfig => setI18nConfig(customI18nConfig));
  }, []);

  return (
    <Provider store={store}>
      <IntlProvider
        locale={i18nConfig.locale}
        key={i18nConfig.locale}
        messages={i18nConfig.messages}
        defaultLocale="en"
      >
        <ConfigLoaderContainer>
          <ConnectedRouter history={history}>
            <Suspense fallback={null}>
              <HeadManager>
                <Layout headerComponent={Header}>
                  <Routes />
                </Layout>
              </HeadManager>
            </Suspense>
          </ConnectedRouter>
        </ConfigLoaderContainer>
      </IntlProvider>
    </Provider>
  );
}

export default Root;
