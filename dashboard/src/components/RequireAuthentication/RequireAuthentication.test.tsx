// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import "@testing-library/jest-dom";
import { screen } from "@testing-library/react";
import { Route, Routes } from "react-router-dom";
import { renderWithProviders } from "shared/specs/mountWrapper";
import { RequireAuthentication } from "./RequireAuthentication";

it("redirects to the /login route if not authenticated", () => {
  renderWithProviders(
    <Routes>
      <Route
        path="/"
        element={
          <RequireAuthentication>
            <h1>Authenticated</h1>
          </RequireAuthentication>
        }
      />
      <Route path="/login" element={<h1>Login</h1>} />
    </Routes>,
    {
      preloadedState: {
        auth: {
          authenticated: false,
          sessionExpired: false,
        },
      },
    },
  );

  expect(screen.getByRole("heading")).toHaveTextContent("Login");
});

it("renders the given component when authenticated", () => {
  renderWithProviders(
    <Routes>
      <Route
        path="/"
        element={
          <RequireAuthentication>
            <h1>Authenticated</h1>
          </RequireAuthentication>
        }
      />
      <Route path="/login" element={<h1>Login</h1>} />
    </Routes>,
    {
      preloadedState: {
        auth: {
          authenticated: true,
          sessionExpired: false,
        },
      },
    },
  );

  expect(screen.getByRole("heading")).toHaveTextContent("Authenticated");
});

it("renders modal to reload the page if the session is expired", () => {
  renderWithProviders(
    <Routes>
      <Route
        path="/"
        element={
          <RequireAuthentication>
            <h1>Authenticated</h1>
          </RequireAuthentication>
        }
      />
      <Route path="/login" element={<h1>Login</h1>} />
    </Routes>,
    {
      preloadedState: {
        auth: {
          authenticated: false,
          sessionExpired: true,
        },
      },
    },
  );

  expect(screen.getByRole("dialog")).toHaveTextContent("Your session has expired");
});
