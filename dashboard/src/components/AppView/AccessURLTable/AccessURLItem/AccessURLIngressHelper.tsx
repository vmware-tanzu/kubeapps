import { IIngressSpec, IIngressTLS, IResource } from "shared/types";
import { IURLItem } from "./IURLItem";

const isURLRegex =
  /^(https?:\/\/)?((([a-z\d]([a-z\d-]*[a-z\d])*)\.)+[a-z-]{2,}|((\d{1,3}\.){3}\d{1,3}))(:\d+)?(\/[-a-z\d%_.~+]*)*(\?[;&a-z\d%_.~+=-]*)?(#[-a-z\d_]*)?$/i;

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

export function IsURL(url: string): boolean {
  return isURLRegex.test(url);
}
