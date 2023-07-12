// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { createMemoryHistory } from "history";
import { Route, Router } from "react-router-dom";
import { renderWithProviders } from "shared/specs/mountWrapper";
import { RequireAuthentication } from "./RequireAuthentication";
import { screen } from "@testing-library/react";
import "@testing-library/jest-dom/extend-expect";

it("redirects to the /login route if not authenticated", () => {
  renderWithProviders((
    <Router history={createMemoryHistory()}>
      <RequireAuthentication>
        <h1>Authenticated</h1>
      </RequireAuthentication>
      <Route path="/login">
        <h1>Login</h1>
      </Route>
    </Router>
  ), {
    preloadedState: {
      auth: {
        authenticated: false,
        sessionExpired: false,
      }
    }
  });

  expect(screen.getByRole("heading")).toHaveTextContent("Login");
});

it("renders the given component when authenticated", () => {
  renderWithProviders((
    <Router history={createMemoryHistory()}>
      <RequireAuthentication>
        <h1>Authenticated</h1>
      </RequireAuthentication>
      <Route path="/login">
        <h1>Login</h1>
      </Route>
    </Router>
  ), {
    preloadedState: {
      auth: {
        authenticated: true,
        sessionExpired: false,
      }
    }
  });

  expect(screen.getByRole("heading")).toHaveTextContent("Authenticated");
});

it("renders modal to reload the page if the session is expired", () => {
  renderWithProviders((
    <Router history={createMemoryHistory()}>
      <RequireAuthentication>
        <h1>Authenticated</h1>
      </RequireAuthentication>
      <Route path="/login">
        <h1>Login</h1>
      </Route>
    </Router>
  ), {
    preloadedState: {
      auth: {
        authenticated: false,
        sessionExpired: true,
      }
    }
  });

  expect(screen.getByRole("dialog")).toHaveTextContent("Your session has expired");
});
