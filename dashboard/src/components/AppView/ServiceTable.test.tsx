import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import ServiceItem from "./ServiceItem";
import ServiceTable from "./ServiceTable";

it("renders a message if there are no services or ingresses", () => {
  const wrapper = shallow(<ServiceTable services={[]} />);
  expect(wrapper.find(ServiceItem)).not.toExist();
  expect(wrapper.text()).toContain("The current application does not contain any service");
});

it("renders a table with a service with a LoadBalancer", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "LoadBalancer",
    },
  } as IResource;
  const services = [service];
  const wrapper = shallow(<ServiceTable services={services} />);
  expect(wrapper.find(ServiceItem).props()).toMatchObject({
    service: { metadata: { name: "foo" }, spec: { type: "LoadBalancer" } },
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
  const services = [service1, service2];
  const wrapper = shallow(<ServiceTable services={services} />);
  expect(
    wrapper
      .find(ServiceItem)
      .at(0)
      .props(),
  ).toMatchObject({ service: service1 });
  expect(
    wrapper
      .find(ServiceItem)
      .at(1)
      .props(),
  ).toMatchObject({ service: service2 });
});
