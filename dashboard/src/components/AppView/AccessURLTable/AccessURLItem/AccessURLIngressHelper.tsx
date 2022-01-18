// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { IIngressSpec, IIngressTLS, IResource } from "shared/types";
import { IURLItem } from "./IURLItem";

// URLs returns the list of URLs obtained from the service status
function URLs(ingress: IResource): string[] {
  const spec = ingress.spec as IIngressSpec;
  const res: string[] = [];
  if (spec.rules) {
    spec.rules.forEach(r => {
      if (r.http && r.http.paths.length > 0) {
        r.http.paths.forEach(p => {
          res.push(getURL(r.host, spec.tls, p.path));
        });
      } else {
        res.push(getURL(r.host, spec.tls));
      }
    });
  } else {
    if (spec.backend && ingress.status?.loadBalancer.ingress[0]) {
      res.push(getURL(ingress.status?.loadBalancer.ingress[0].ip, spec.tls));
    }
  }
  return res;
}

// getURL returns a full URL based on a hostname, a TLS configuration and a optional path
function getURL(hostname: string, tls?: IIngressTLS[], path?: string) {
  // If the hostname is configured within the TLS hosts it will use HTTPS
  const protocol =
    tls &&
    tls.some(tlsRule => {
      return tlsRule.hosts && tlsRule.hosts.indexOf(hostname) > -1;
    })
      ? "https"
      : "http";
  return `${protocol}://${hostname}${path || ""}`;
}

export function GetURLItemFromIngress(ingress: IResource) {
  return {
    name: ingress.metadata.name,
    type: "Ingress",
    isLink: true,
    URLs: URLs(ingress),
  } as IURLItem;
}

// ShouldGenerateLink returns true if the string is, likely, a URL and
// it, presumably, does not contain a regular expression.
// For instance:
//   http://wordpress.local/example-good should be true
//   http://wordpress.local/example-bad(/|$)(.*) should be false
export function ShouldGenerateLink(url: string): boolean {
  // Check if it is a valid URL, delegating the check to the browser API
  let builtURL;
  try {
    // Note this browser API constructor will accept many invalid URLs (not well-encoded, wrong IP octets, etc.)
    builtURL = new URL(url);
  } catch (_) {
    return false;
  }
  if (!(builtURL.protocol === "http:" || builtURL.protocol === "https:")) {
    return false;
  }

  // Check if any url fragment includes a character likely used in a regex, that is, is not encoded
  // c.f. https://datatracker.ietf.org/doc/html/rfc3986#section-2.2
  // even if it won't catch every possible case, we avoid having an O(x^n) regex
  const regex = /^[^:/?#[\]@!$&'*+,;=]*$/;

  // If it is an URL but it includes some "ilegal" characters, then return false
  return !builtURL.pathname.split("/")?.some(p => !regex.test(p));
}
