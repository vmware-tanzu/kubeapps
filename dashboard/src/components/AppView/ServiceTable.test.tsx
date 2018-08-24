import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import ServiceItem from "./ServiceItem";
import ServiceTable from "./ServiceTable";

it("renders a a table with a LoadBalancer", () => {
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

it("renders a a table with a LoadBalancer", () => {
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
