import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import ServiceItem from "./ServiceItem";
import ServiceTable from "./ServiceTable";

it("renders a table with a service with a LoadBalancer", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "LoadBalancer",
    },
  } as IResource;
  const services = new Map<string, IResource>();
  services[service.metadata.name] = service;
  const wrapper = shallow(<ServiceTable services={services} extended={false} />);
  expect(wrapper.find(ServiceItem).props()).toMatchObject({
    extended: false,
    service: { metadata: { name: "foo" }, spec: { type: "LoadBalancer" } },
  });
});

it("renders a table with a service with a ClusterIP", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "ClusterIP",
    },
  } as IResource;
  const services = new Map<string, IResource>();
  services[service.metadata.name] = service;
  const wrapper = shallow(<ServiceTable services={services} extended={true} />);
  expect(wrapper.find(ServiceItem).props()).toMatchObject({
    extended: true,
    service: { metadata: { name: "foo" }, spec: { type: "ClusterIP" } },
  });
});

it("renders a table with a service with two services", () => {
  const service1 = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "LoadBalancer",
    },
  } as IResource;
  const service2 = {
    metadata: {
      name: "bar",
    },
    spec: {
      type: "ClusterIP",
    },
  } as IResource;
  const services = new Map<string, IResource>();
  services[service1.metadata.name] = service1;
  services[service2.metadata.name] = service2;
  const wrapper = shallow(<ServiceTable services={services} extended={true} />);
  expect(
    wrapper
      .find(ServiceItem)
      .at(0)
      .props(),
  ).toMatchObject({ extended: true, service: service1 });
  expect(
    wrapper
      .find(ServiceItem)
      .at(1)
      .props(),
  ).toMatchObject({ extended: true, service: service2 });
});
