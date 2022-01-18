// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { IHTTPIngressPath, IIngressRule, IResource } from "shared/types";
import { GetURLItemFromIngress, ShouldGenerateLink } from "./AccessURLIngressHelper";

describe("GetURLItemFromIngress", () => {
  interface ITest {
    description: string;
    hosts?: string[];
    paths?: IHTTPIngressPath[];
    tlsHosts?: string[];
    expectedURLs: string[];
    status?: any;
    backend?: any;
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
    {
      description: "it should add an ingress with a default backend",
      status: {
        loadBalancer: {
          ingress: [{ ip: "1.2.3.4" }],
        },
      },
      backend: {},
      expectedURLs: ["http://1.2.3.4"],
    },
  ];
  tests.forEach(test => {
    it(test.description, () => {
      const ingress = {
        metadata: {
          name: "foo",
        },
        spec: {
          backend: test.backend,
        },
        status: test.status,
      } as IResource;
      if (test.hosts) {
        ingress.spec.rules = [];
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

describe("IsURL", () => {
  interface ITest {
    description: string;
    fullURL: string;
  }
  // We are using some of the examples from the https://mathiasbynens.be/demo/url-regex list
  describe("Should return true for valid URLs", () => {
    const validURLs: ITest[] = [
      {
        description: "it should return true to a simple url",
        fullURL: "http://wordpress.local/example-test",
      },
      {
        description: "it should return true to a simple url (https)",
        fullURL: "https://wordpress.local/example-test",
      },
      {
        description: "it should return true to a simple url (host is ip)",
        fullURL: "http://1.1.1.1/example-test",
      },
      {
        description: "it should return true to a url with an emoji",
        fullURL: "http://wordpress.local/♨️",
      },
      {
        description: "it should return true to a url with a idn domain",
        fullURL: "http://wordpress.♡.com",
      },
      {
        description: "it should return true to a url with an emoji",
        fullURL: "http://wordpress.local/♨️",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.com/blah_blah",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.com/blah_blah_(wikipedia)",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.com/blah_blah_(wikipedia)_(again)",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://www.example.com/wpstyle/?p=364",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "https://www.example.com/foo/?bar=baz&inga=42&quux",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://✪df.ws/123",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://userid:password@example.com:8080",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://userid:password@example.com:8080/",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://userid@example.com",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://userid@example.com/",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://userid@example.com:8080",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://userid@example.com:8080/",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://userid:password@example.com",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://userid:password@example.com/",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://142.42.1.1/",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://142.42.1.1:8080/",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://➡.ws/䨹",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://⌘.ws",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://⌘.ws/",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.com/blah_(wikipedia)#cite-1",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.com/blah_(wikipedia)_blah#cite-1",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.com/unicode_(✪)_in_parens",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.com/(something)?after=parens",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://☺.damowmow.com/",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://code.google.com/events/#&product=browser",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://j.mp",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.bar/baz",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://foo.bar/?q=Test%20URL-encoded%20stuff",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://مثال.إختبار",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://例子.测试",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://उदाहरण.परीक्षा",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://-.~_!$&'()*+,;=:%40:80%2f::::::@example.com",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://1337.net",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://a.b-c.de",
      },
      {
        description: "it should return true to a mathiasbynens's subset valid item",
        fullURL: "http://223.255.255.254",
      },
    ];
    validURLs.forEach(test => {
      it(`${test.description}: ${test.fullURL}`, () => {
        expect(ShouldGenerateLink(test.fullURL)).toBe(true);
      });
    });
  });

  describe("Should return false for invalid URLs", () => {
    const invalidURLs: ITest[] = [
      {
        description: "it should return false to a 'rfc-ilegal' url with a regex",
        fullURL: "http://wordpress.local/example-bad(/|$)(.*)",
      },
      {
        description: "it should return false to a 'rfc-legal' url with a regex",
        fullURL: "http://wordpress.local/example-bad/[a-z]+",
      },

      {
        description: "it should return false to a 'rfc-ilegal' url with a regex (https)",
        fullURL: "https://wordpress.local/example-bad(/|$)(.*)",
      },
      {
        description: "it should return false to a 'rfc-legal' url with a regex (https)",
        fullURL: "https://wordpress.local/example-bad/[a-z]+",
      },

      {
        description: "it should return false to a 'rfc-ilegal' url with a regex (host is ip)",
        fullURL: "http://1.1.1.1/example-bad(/|$)(.*)",
      },
      {
        description: "it should return false to a 'rfc-legal' url with a regex (host is ip)",
        fullURL: "http://1.1.1.1/example-bad/[a-z]+",
      },
      {
        description: "it should return false to a 'rfc-legal' url with a regex",
        fullURL: "http://wordpress.local/example-bad/*",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "http://",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "http://?",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "http://??",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "http://??/",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "http://#",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "http://##",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "http://##/",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "//",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "//a",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "///a",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "///",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "foo.com",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "rdar://1234",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "h://test",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: "http:// shouldfail.com",
      },
      {
        description: "it should return false to a mathiasbynens's subset invalid item",
        fullURL: ":// should fail",
      },
    ];
    invalidURLs.forEach(test => {
      it(`${test.description}: ${test.fullURL}`, () => {
        expect(ShouldGenerateLink(test.fullURL)).toBe(false);
      });
    });
  });
});
