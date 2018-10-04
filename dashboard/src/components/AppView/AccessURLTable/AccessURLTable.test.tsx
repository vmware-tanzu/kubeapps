import { shallow } from "enzyme";
import * as React from "react";

import { IResource, IServiceSpec, IServiceStatus } from "shared/types";
import AccessURLItem from "./AccessURLItem";
import AccessURLTable from "./AccessURLTable";

it("should omit the Service Table if there are no public services", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "ClusterIP",
      ports: [{ port: 8080 }],
    } as IServiceSpec,
    status: {
      loadBalancer: {},
    } as IServiceStatus,
  } as IResource;
  const services = {};
  services[service.metadata.name] = service;
  const wrapper = shallow(<AccessURLTable services={services} />);
  expect(wrapper.text()).toBe("");
});

it("should show the table if any service is a LoadBalancer", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "LoadBalancer",
      ports: [{ port: 8080 }],
    } as IServiceSpec,
    status: {
      loadBalancer: {},
    } as IServiceStatus,
  } as IResource;
  const services = {};
  services[service.metadata.name] = service;
  const wrapper = shallow(<AccessURLTable services={services} />);
  expect(wrapper.find(AccessURLItem)).toExist();
  expect(wrapper).toMatchSnapshot();
});
