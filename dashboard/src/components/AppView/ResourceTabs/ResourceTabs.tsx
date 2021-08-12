import Tabs from "components/Tabs";
import ResourceTable from "containers/ResourceTableContainer";
import ResourceRef from "shared/ResourceRef";

interface IAppViewResourceRefs {
  deployments: ResourceRef[];
  statefulsets: ResourceRef[];
  daemonsets: ResourceRef[];
  services: ResourceRef[];
  secrets: ResourceRef[];
  otherResources: ResourceRef[];
}

export default function ResourceTabs({
  deployments,
  statefulsets,
  daemonsets,
  secrets,
  services,
  otherResources,
}: IAppViewResourceRefs) {
  const columns = [];
  const data = [];
  if (deployments.length) {
    columns.push("Deployments");
    data.push(<ResourceTable resourceRefs={deployments} key="deployments" id="deployments" />);
  }
  if (statefulsets.length) {
    columns.push("StatefulSets");
    data.push(<ResourceTable resourceRefs={statefulsets} key="statefulsets" id="statefulsets" />);
  }
  if (daemonsets.length) {
    columns.push("DaemonSets");
    data.push(<ResourceTable resourceRefs={daemonsets} key="daemonsets" id="daemonsets" />);
  }
  if (secrets.length) {
    columns.push("Secrets");
    data.push(<ResourceTable resourceRefs={secrets} key="secrets" id="secrets" />);
  }
  if (services.length) {
    columns.push("Services");
    data.push(<ResourceTable resourceRefs={services} key="services" id="services" />);
  }
  if (otherResources.length) {
    columns.push("Other Resources");
    data.push(
      <ResourceTable resourceRefs={otherResources} key="otherResources" id="otherResources" />,
    );
  }
  // TODO(agamez): temporary fix since the API is not returning any resources yet
  const hasResources =
    deployments.length > 0 ||
    statefulsets.length > 0 ||
    daemonsets.length > 0 ||
    services.length > 0 ||
    secrets.length > 0 ||
    otherResources.length > 0;
  return (
    <>
      {hasResources && (
        <section aria-labelledby="resources-table">
          <h5 className="section-title" id="resources-table">
            Application Resources
          </h5>
          <Tabs id="resource-table-tabs" columns={columns} data={data} />
        </section>
      )}
    </>
  );
}
