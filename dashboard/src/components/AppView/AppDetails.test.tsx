import { shallow } from "enzyme";
import * as React from "react";
import { IResource } from "../../shared/types";
import DeploymentTable from "./DeploymentTable";
import ServiceTable from "./ServiceTable";

import AppDetails from "./AppDetails";

const defaultProps = {
  deployments: {},
  otherResources: {},
  services: {},
};

it("renders a deployment details", () => {
  const deployments = {};
  const dep = "foo";
  deployments[dep] = {
    kind: "Deployment",
    metadata: {
      name: "foo",
    },
    status: {},
  } as IResource;
  const wrapper = shallow(<AppDetails {...defaultProps} deployments={deployments} />);
  expect(wrapper.find(DeploymentTable).props().deployments).toMatchObject(deployments);
});

it("renders a services details", () => {
  const services = {};
  const svc = "foo";
  services[svc] = {
    kind: "Service",
    metadata: {
      name: "foo",
    },
    spec: {
      ports: [],
      type: "ClusterIP",
    },
    status: {},
  } as IResource;
  const wrapper = shallow(<AppDetails {...defaultProps} services={services} />);
  expect(wrapper.find(ServiceTable).props().services).toMatchObject(services);
});

it("renders a other resources details", () => {
  const otherResources = {};
  const r1 = "foo";
  otherResources[r1] = {
    kind: "Secret",
    metadata: {
      name: "foo",
    },
    status: {},
  } as IResource;
  const r2 = "bar";
  otherResources[r2] = {
    kind: "PersistentVolume",
    metadata: {
      name: "foo",
    },
    status: {},
  } as IResource;
  const wrapper = shallow(<AppDetails {...defaultProps} otherResources={otherResources} />);
  expect(wrapper).toMatchSnapshot();
});
