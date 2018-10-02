import { shallow } from "enzyme";
import * as React from "react";
import { IResource } from "../../shared/types";
import DeploymentTable from "./DeploymentTable";
import ServiceTable from "./ServiceTable";

import AppDetails from "./AppDetails";

const defaultProps = {
  deployments: {} as Map<string, IResource>,
  otherResources: {} as Map<string, IResource>,
  services: {} as Map<string, IResource>,
};

it("renders a deployment details", () => {
  const deployments = new Map<string, IResource>();
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
  const services = new Map<string, IResource>();
  const svc = "foo";
  services.set(svc, {
    kind: "Service",
    metadata: {
      name: "foo",
    },
    spec: {
      ports: [],
      type: "ClusterIP",
    },
    status: {},
  } as IResource);
  const wrapper = shallow(<AppDetails {...defaultProps} services={services} />);
  expect(wrapper.find(ServiceTable).props().services).toMatchObject(services);
});

it("renders a other resources details", () => {
  const otherResources = new Map<string, IResource>();
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
