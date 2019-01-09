import { shallow } from "enzyme";
import * as React from "react";

import { IKubeItem, IResource, IResourceRef } from "../../../shared/types";
import ServiceItem from "./ServiceItem";

const serviceRef = {
  name: "foo",
} as IResourceRef;

const kubeItem: IKubeItem<IResource> = {
  isFetching: false,
};

it("renders a simple view without IP", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "ClusterIP",
      ports: [],
    },
  } as IResource;
  kubeItem.item = service;
  const wrapper = shallow(
    <ServiceItem service={kubeItem} serviceRef={serviceRef} getService={jest.fn()} />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("renders a simple view with IP", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      ports: [{ port: 80 }],
      type: "LoadBalancer",
    },
    status: {
      loadBalancer: {
        ingress: [{ ip: "1.2.3.4" }],
      },
    },
  } as IResource;
  kubeItem.item = service;
  const wrapper = shallow(
    <ServiceItem service={kubeItem} serviceRef={serviceRef} getService={jest.fn()} />,
  );
  expect(wrapper.text()).toContain("1.2.3.4");
});

it("renders a view with IP", () => {
  const service = {
    metadata: { name: "foo" },
    spec: { ports: [{ port: 80 }], type: "LoadBalancer" },
    status: { loadBalancer: { ingress: [{ ip: "1.2.3.4" }] } },
  } as IResource;
  kubeItem.item = service;
  const wrapper = shallow(
    <ServiceItem service={kubeItem} serviceRef={serviceRef} getService={jest.fn()} />,
  );
  expect(wrapper).toMatchSnapshot();
});
