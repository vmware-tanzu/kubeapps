import { get } from "lodash";
import { useEffect } from "react";

import actions from "actions";
import Table from "components/js/Table";
import Tooltip from "components/js/Tooltip";
import { useDispatch, useSelector } from "react-redux";
import ResourceRef from "shared/ResourceRef";
import { flattenResources } from "shared/utils";
import LoadingWrapper from "../../../components/LoadingWrapper/LoadingWrapper";
import { IK8sList, IKubeItem, IResource, IServiceSpec, IStoreState } from "../../../shared/types";
import isSomeResourceLoading from "../helpers";
import { GetURLItemFromIngress, IsURL } from "./AccessURLItem/AccessURLIngressHelper";
import { GetURLItemFromService } from "./AccessURLItem/AccessURLServiceHelper";
import "./AccessURLTable.css";

interface IAccessURLTableProps {
  ingressRefs: ResourceRef[];
  serviceRefs: ResourceRef[];
}

function elemHasItems(i: IKubeItem<IResource | IK8sList<IResource, {}>>) {
  if (i.error) {
    return true;
  }
  if (i.item) {
    const list = i.item as IK8sList<IResource, {}>;
    if (list.items && list.items.length === 0) {
      return false;
    }
    return true;
  }
  return false;
}

function hasItems(svcs: Array<IKubeItem<IResource>>, ingresses: Array<IKubeItem<IResource>>) {
  return (
    (svcs.length && svcs.some(svc => elemHasItems(svc))) ||
    (ingresses.length && ingresses.some(ingress => elemHasItems(ingress)))
  );
}

function filterPublicServices(services: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>) {
  const result: Array<IKubeItem<IResource>> = [];
  services.forEach(s => {
    if (s.item) {
      const listItem = s.item as IK8sList<IResource, {}>;
      if (listItem.items) {
        listItem.items.forEach(item => {
          if (item.spec.type === "LoadBalancer") {
            result.push({ isFetching: false, item });
          }
        });
      } else {
        const spec = (s.item as IResource).spec as IServiceSpec;
        if (spec.type === "LoadBalancer") {
          result.push(s as IKubeItem<IResource>);
        }
      }
    }
  });
  return result;
}

function flattenIngresses(ingresses: Array<IKubeItem<IResource | IK8sList<IResource, {}>>>) {
  const result: Array<IKubeItem<IResource>> = [];
  ingresses.forEach(ingress => {
    const list = ingress.item as IK8sList<IResource, {}>;
    if (list && list.items) {
      list.items.forEach(i => {
        result.push({ isFetching: false, item: i, error: ingress.error });
      });
    } else {
      result.push(ingress as IKubeItem<IResource>);
    }
  });
  return result;
}

function getAnchors(URLs: string[]) {
  return URLs.map(URL => getAnchor(URL));
}

function getAnchor(URL: string) {
  return (
    <div className="margin-b-sm" key={URL}>
      <a href={URL} target="_blank" rel="noopener noreferrer">
        {URL}
      </a>
    </div>
  );
}

function getSpan(URL: string) {
  return (
    <div className="margin-b-sm" key={URL}>
      <span>{URL}</span>
    </div>
  );
}

function getUnknown(key: string) {
  return (
    <div className="margin-b-sm" key={key}>
      <span>Unknown</span>
    </div>
  );
}

function getNotes(resource?: IResource) {
  if (!resource) {
    return getUnknown("unknown-notes");
  }
  const ips: Array<{ ip: string }> = get(resource, "status.loadBalancer.ingress", []);
  if (ips.length) {
    return <span>IP(s): {ips.map(ip => ip.ip).join(", ")}</span>;
  }
  return (
    <span className="tooltip-wrapper">
      Not associated with any IP.{" "}
      <Tooltip
        label="pending-tooltip"
        id={`${resource.metadata.name}-pending-tooltip`}
        icon="help"
        position="bottom-left"
        large={true}
        iconProps={{ solid: true, size: "sm" }}
      >
        Depending on your cloud provider of choice, it may take some time for an access URL to be
        available for the application and the Service will stay in a "Pending" state until a URL is
        assigned. If using Minikube, you will need to run <code>minikube tunnel</code> in your
        terminal in order for an IP address to be assigned to your application.
      </Tooltip>
    </span>
  );
}

export default function AccessURLTable({ ingressRefs, serviceRefs }: IAccessURLTableProps) {
  const dispatch = useDispatch();

  useEffect(() => {
    // Fetch all related Ingress resources. We don't need to fetch Services as
    // they are expected to be watched by the ServiceTable.
    ingressRefs.forEach(r => dispatch(actions.kube.getResource(r)));
  }, [dispatch, ingressRefs]);
  const ingresses = useSelector((state: IStoreState) =>
    flattenResources(ingressRefs, state.kube.items),
  ) as Array<IKubeItem<IResource>>;
  const services = useSelector((state: IStoreState) =>
    flattenResources(serviceRefs, state.kube.items),
  ) as Array<IKubeItem<IResource>>;

  if (isSomeResourceLoading(ingresses.concat(services))) {
    return (
      <section aria-labelledby="access-urls-title">
        <h5 className="section-title" id="access-urls-title">
          Access URLs
        </h5>
        <LoadingWrapper loaded={false} />
      </section>
    );
  }
  if (!hasItems(services, ingresses)) {
    return null;
  }

  let result = <p>The current application does not expose a public URL.</p>;
  const publicServices = filterPublicServices(services);
  if (publicServices.length > 0 || ingresses.length > 0) {
    const columns = [
      {
        accessor: "url",
        Header: "URL",
      },
      {
        accessor: "type",
        Header: "Type",
      },
      {
        accessor: "notes",
        Header: "Notes",
      },
    ];
    const allIngresses = flattenIngresses(ingresses);
    const data = publicServices
      .map(svc => {
        const urlItem = GetURLItemFromService(svc.item);
        return {
          url: urlItem.isLink ? getAnchors(urlItem.URLs) : urlItem.URLs.join(","),
          type: urlItem.type,
          notes: svc.error ? <span>Error: {svc.error.message}</span> : getNotes(svc.item),
        };
      })
      .concat(
        allIngresses.map((ingress, index) => {
          return {
            url: ingress.item
              ? GetURLItemFromIngress(ingress.item).URLs.map(
                  // check whether each URL is, indeed, a valid URL.
                  // If so, render the <a>, othersiwe, render a simple <span>
                  url => (IsURL(url) ? getAnchor(url) : getSpan(url)),
                )
              : [getUnknown(index.toString())], // render a simple span with "unknown"
            type: "Ingress",
            notes: ingress.error ? (
              <span>Error: {ingress.error.message}</span>
            ) : (
              getNotes(ingress.item)
            ),
          };
        }),
      );
    result = <Table data={data} columns={columns} />;
  }
  return (
    <section aria-labelledby="access-urls-title">
      <h5 className="section-title" id="access-urls-title">
        Access URLs
      </h5>
      {result}
    </section>
  );
}
