// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { IResource, IServiceStatus } from "shared/types";
import { GetURLItemFromService } from "./AccessURLServiceHelper";

describe("GetURLItemFromService", () => {
  interface ITest {
    description: string;
    ports: any[];
    ingress: any[];
    expectedLink: boolean;
    expectedURLs: string[];
  }
  const tests: ITest[] = [
    {
      description: "it not return a link if the ingress definition is empty",
      ports: [{ port: 8080 }],
      ingress: [],
      expectedLink: false,
      expectedURLs: ["Pending"],
    },
    {
      description: "it should show the IP and port if it's not known",
      ports: [{ port: 8080 }],
      ingress: [{ ip: "1.2.3.4" }],
      expectedLink: true,
      expectedURLs: ["http://1.2.3.4:8080"],
    },
    {
      description: "it should show the hostname and port if it's not known",
      ports: [{ port: 8080 }],
      ingress: [{ hostname: "1.2.3.4" }],
      expectedLink: true,
      expectedURLs: ["http://1.2.3.4:8080"],
    },
    {
      description: "it should show the IP and skip the port if it's known",
      ports: [{ port: 80 }],
      ingress: [{ ip: "1.2.3.4" }],
      expectedLink: true,
      expectedURLs: ["http://1.2.3.4"],
    },
    {
      description: "it should show the https URL if the port is 443",
      ports: [{ port: 443 }],
      ingress: [{ ip: "1.2.3.4" }],
      expectedLink: true,
      expectedURLs: ["https://1.2.3.4"],
    },
    {
      description: "it should show several URLs if there are multiple ports",
      ports: [{ port: 8080 }, { port: 8081 }],
      ingress: [{ ip: "1.2.3.4" }],
      expectedLink: true,
      expectedURLs: ["http://1.2.3.4:8080", "http://1.2.3.4:8081"],
    },
    {
      description: "it should show several URLs if there are ingress ports",
      ports: [{ port: 8080 }, { port: 8081 }],
      ingress: [{ ip: "1.2.3.4" }, { hostname: "foo.bar" }],
      expectedLink: true,
      expectedURLs: [
        "http://1.2.3.4:8080",
        "http://1.2.3.4:8081",
        "http://foo.bar:8080",
        "http://foo.bar:8081",
      ],
    },
  ];
  tests.forEach(test => {
    it(test.description, () => {
      const service = {
        metadata: {
          name: "foo",
        },
        spec: {
          type: "LoadBalancer",
          ports: test.ports,
        },
        status: {
          loadBalancer: {
            ingress: test.ingress,
          },
        } as IServiceStatus,
      } as IResource;
      const svcItem = GetURLItemFromService(service);
      expect(test.expectedLink).toEqual(svcItem.isLink);
      expect(test.expectedURLs).toEqual(svcItem.URLs);
    });
  });
});
