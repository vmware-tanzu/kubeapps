/**
 * Components using the react-intl module require access to the intl context.
 * This is not available when mounting single components in Enzyme.
 * These helper functions aim to address that and wrap a valid,
 * English-locale intl context around them.
 */

import { mount, shallow } from "enzyme";
import * as React from "react";
import { IntlProvider } from "react-intl";
import { Provider } from "react-redux";
import { BrowserRouter as Router } from "react-router-dom";
import { MockStore } from "redux-mock-store";
import I18n from "../I18n";

// Default to english 1i8nconfiguration
const messages = I18n.getDefaultConfig().messages;
const locale = I18n.getDefaultConfig().locale;

export const mountIntl = (node: React.ReactElement) =>
  mount(
    <IntlProvider locale={locale} key={locale} messages={messages} defaultLocale={locale}>
      {node}
    </IntlProvider>,
  );

export const mountWrapperIntl = (store: MockStore, children: React.ReactElement) =>
  mount(
    <Provider store={store}>
      <IntlProvider locale={locale} key={locale} messages={messages} defaultLocale={locale}>
        <Router>{children}</Router>
      </IntlProvider>
    </Provider>,
  );

export const shallowIntl = (node: React.ReactElement) =>
  shallow(
    <IntlProvider locale={locale} key={locale} messages={messages} defaultLocale={locale}>
      {node}
    </IntlProvider>,
  );
