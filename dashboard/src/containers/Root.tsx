import Header from "components/Header";
import Layout from "components/Layout";
import { ConnectedRouter } from "connected-react-router";
import * as React from "react";
import { IntlProvider } from "react-intl";
import { Provider } from "react-redux";

import { getCustomI18nConfig, getDefaulI18nConfig, ISupportedLangs } from "shared/I18n";
import store, { history } from "../store";
// TODO(andresmgot): Containers should be no longer needed, replace them when possible
import ConfigLoaderContainer from "./ConfigLoaderContainer";
import Routes from "./RoutesContainer";

interface IRootState {
  locale: string;
  messages: Record<string, string>;
}
class Root extends React.Component<{}, IRootState> {
  constructor(props: {} | IRootState) {
    super(props);
    this.state = getDefaulI18nConfig();
    this.initLocale();
  }

  public async initLocale() {
    const fullLang = (navigator.languages && navigator.languages[0]) || navigator.language;
    const lang = fullLang.toLowerCase().split(/[_-]+/)[0];
    this.setState(await getCustomI18nConfig(ISupportedLangs[lang]));
  }

  public render() {
    return (
      <Provider store={store}>
        <ConfigLoaderContainer>
          <ConnectedRouter history={history}>
            <React.Suspense fallback={null}>
              <IntlProvider
                locale={this.state.locale}
                key={this.state.locale}
                messages={this.state.messages}
                defaultLocale="en"
              >
                <Layout headerComponent={Header}>
                  <Routes />
                </Layout>
              </IntlProvider>
            </React.Suspense>
          </ConnectedRouter>
        </ConfigLoaderContainer>
      </Provider>
    );
  }
}

export default Root;
