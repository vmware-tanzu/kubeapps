import { IHTTPIngressPath, IIngressRule, IIngressSpec, IResource } from "shared/types";
import { GetURLItemFromIngress } from "./AccessURLIngressHelper";

describe("GetURLItemFromIngress", () => {
  interface ITest {
    description: string;
    hosts?: string[];
    paths?: IHTTPIngressPath[];
    tlsHosts?: string[];
    expectedURLs: string[];
  }
  const tests: ITest[] = [
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
    {
      description: "it should ignore an ingress if it has no hosts",
      hosts: undefined,
      paths: [],
      tlsHosts: ["foo.bar"],
      expectedURLs: [],
    },
    {
      description: "it should ignore paths if undefined",
      hosts: ["foo.bar"],
      paths: undefined,
      tlsHosts: [],
      expectedURLs: ["http://foo.bar"],
    },
    {
      description: "it should ignore TLS if the hosts are undefined",
      hosts: ["foo.bar"],
      paths: [],
      tlsHosts: undefined,
      expectedURLs: ["http://foo.bar"],
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
      if (test.hosts) {
        test.hosts.forEach(h => {
          const rule = {
            host: h,
            http: {
              paths: [],
            },
          } as IIngressRule;
          if (test.paths && test.paths.length > 0) {
            rule.http.paths = test.paths;
          }
          ingress.spec.rules.push(rule);
        });
      }
      if (test.tlsHosts && test.tlsHosts.length > 0) {
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
