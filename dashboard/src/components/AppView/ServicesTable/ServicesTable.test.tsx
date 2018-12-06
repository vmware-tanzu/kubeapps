import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";

import LoadingWrapper from "../../../components/LoadingWrapper";
import { IResource } from "../../../shared/types";
import ServiceItem from "./ServiceItem";
import ServiceTable from "./ServicesTable";

context("when fetching ingresses or services", () => {
  itBehavesLike("aLoadingComponent", {
    component: ServiceTable,
    props: {
      services: [{ isFetching: true }],
    },
  });
});

it("renders a message if there are no services or ingresses", () => {
  const wrapper = shallow(<ServiceTable services={[]} />);
  expect(
    wrapper
      .find(LoadingWrapper)
      .shallow()
      .find(ServiceItem),
  ).not.toExist();
  expect(
    wrapper
      .find(LoadingWrapper)
      .shallow()
      .text(),
  ).toContain("The current application does not contain any service");
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
  const services = [{ isFetching: false, item: service }];
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
  const services = [{ isFetching: false, item: service1 }, { isFetching: false, item: service2 }];
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
