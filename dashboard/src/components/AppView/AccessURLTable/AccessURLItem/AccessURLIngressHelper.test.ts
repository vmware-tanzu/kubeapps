import { IHTTPIngressPath, IIngressRule, IIngressSpec, IResource } from "shared/types";
import { GetURLItemFromIngress } from "./AccessURLIngressHelper";

describe("GetURLItemFromIngress", () => {
  interface Itest {
    description: string;
    hosts: string[];
    paths: IHTTPIngressPath[];
    tlsHosts: string[];
    expectedURLs: string[];
  }
  const tests: Itest[] = [
    {
      description: "it should show the host without port",
      hosts: ["foo.bar"],
      paths: [],
      tlsHosts: [],
      expectedURLs: ["http://foo.bar"],
    },
    {
      description: "it should show several hosts without port",
      hosts: ["foo.bar", "not-foo.bar"],
      paths: [],
      tlsHosts: [],
      expectedURLs: ["http://foo.bar", "http://not-foo.bar"],
    },
    {
      description: "it should show the host with the different paths",
      hosts: ["foo.bar"],
      paths: [{ path: "/one" }, { path: "/two" }],
      tlsHosts: [],
      expectedURLs: ["http://foo.bar/one", "http://foo.bar/two"],
    },
    {
      description: "it should show TLS hosts with https",
      hosts: ["foo.bar", "not-foo.bar"],
      paths: [],
      tlsHosts: ["foo.bar"],
      expectedURLs: ["https://foo.bar", "http://not-foo.bar"],
    },
  ];
  tests.forEach(test => {
    it(test.description, () => {
      const ingress = {
        metadata: {
          name: "foo",
        },
        spec: {
          rules: [],
        } as IIngressSpec,
      } as IResource;
      test.hosts.forEach(h => {
        const rule = {
          host: h,
          http: {
            paths: [],
          },
        } as IIngressRule;
        if (test.paths.length > 0) {
          rule.http.paths = test.paths;
        }
        ingress.spec.rules.push(rule);
      });
      if (test.tlsHosts.length > 0) {
        ingress.spec.tls = [
          {
            hosts: test.tlsHosts,
          },
        ];
      }
      const ingressItem = GetURLItemFromIngress(ingress);
      expect(ingressItem.isLink).toBe(true);
      expect(ingressItem.URLs).toEqual(test.expectedURLs);
    });
  });
});
