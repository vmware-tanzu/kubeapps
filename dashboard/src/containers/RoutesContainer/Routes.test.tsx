// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import "@testing-library/jest-dom";
import { screen } from "@testing-library/react";
import { IAuthState } from "reducers/auth";
import { IClusterState } from "reducers/cluster";
import { renderWithProviders } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import AppRoutes from "./Routes";

// Mocking SwaggerUI to a simple empty <div> to prevent issues with Jest
jest.mock("swagger-ui-react", () => {
  return {
    SwaggerUI: () => <div />,
  };
});

it("invalid path should show a 404 error", () => {
  renderWithProviders(<AppRoutes />, {
    preloadedState: {
      auth: {
        authenticated: true,
      },
    },
    initialEntries: ["/random"],
  });

  expect(screen.getByRole("heading")).toHaveTextContent(
    "The page you are looking for can't be found.",
  );
});

it("should render a redirect to the default cluster and namespace", () => {
  renderWithProviders(<AppRoutes />, {
    preloadedState: {
      auth: {
        authenticated: true,
      },
      clusters: {
        currentCluster: "default",
        clusters: {
          default: {
            currentNamespace: "default",
          } as Partial<IClusterState>,
        },
      },
    },
    initialEntries: ["/"],
  });

  expect(screen.getByRole("heading")).toHaveTextContent("Applications");
  expect(screen.getByRole("link", { name: "Deploy" })).toHaveAttribute(
    "href",
    "/c/default/ns/default/catalog",
  );
});

it("should render a redirect to the login page", () => {
  renderWithProviders(<AppRoutes />, {
    preloadedState: {
      auth: {
        authenticated: false,
      },
    },
    initialEntries: ["/"],
  });

  expect(screen.getByLabelText("Token")).toHaveAttribute("placeholder", "Paste token here");
});

it("should render a redirect to the login page (even with cluster or ns info)", () => {
  renderWithProviders(<AppRoutes />, {
    preloadedState: {
      auth: {
        authenticated: false,
      } as Partial<IAuthState>,
      clusters: {
        currentCluster: "default",
        clusters: {
          default: {
            currentNamespace: "default",
          } as Partial<IClusterState>,
        },
      },
    } as Partial<IStoreState>,
    initialEntries: ["/"],
  });

  expect(screen.getByLabelText("Token")).toHaveAttribute("placeholder", "Paste token here");
});

it("should render a loading wrapper if authenticated but the cluster and ns info is not populated", () => {
  renderWithProviders(<AppRoutes />, {
    preloadedState: {
      auth: {
        authenticated: true,
      },
    },
    initialEntries: ["/"],
  });

  expect(screen.getByText("Fetching Cluster Info...")).toBeInTheDocument();
});

it("should render a warning message if operators are deactivated", () => {
  renderWithProviders(<AppRoutes />, {
    preloadedState: {
      auth: {
        authenticated: true,
      },
      config: {
        featureFlags: {
          operators: false,
        },
      },
    },
    initialEntries: [`${app.config.operators("default", "default")}/some/path`],
  });
  expect(
    screen.getByText("Operators support has been deactivated", { exact: false }),
  ).toBeInTheDocument();
});

it("should route to operators if enabled", () => {
  renderWithProviders(<AppRoutes />, {
    preloadedState: {
      auth: {
        authenticated: true,
      },
      config: {
        featureFlags: {
          operators: true,
        },
      },
    },
    initialEntries: [`${app.config.operators("default", "default")}`],
  });

  expect(screen.getByText("Fetching Operators...")).toBeInTheDocument();
});
