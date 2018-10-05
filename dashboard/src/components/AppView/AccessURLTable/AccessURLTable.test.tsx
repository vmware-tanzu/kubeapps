import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { IIngressSpec, IResource, IServiceSpec, IServiceStatus } from "shared/types";
import AccessURLItem from "./AccessURLItem";
import AccessURLTable from "./AccessURLTable";

context("when the app contain services", () => {
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
    const wrapper = shallow(<AccessURLTable services={services} ingresses={{}} />);
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
    const wrapper = shallow(<AccessURLTable services={services} ingresses={{}} />);
    expect(wrapper.find(AccessURLItem)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});

context("when the app contain ingresses", () => {
  it("should show the table with available ingresses", () => {
    const ingress = {
      metadata: {
        name: "foo",
      },
      spec: {
        rules: [
          {
            host: "foo.bar",
            http: {
              paths: [{ path: "/ready" }],
            },
          },
        ],
      } as IIngressSpec,
    };
    const ingresses = {};
    ingresses[ingress.metadata.name] = ingress;
    const wrapper = shallow(<AccessURLTable services={{}} ingresses={ingresses} />);
    expect(wrapper.find(AccessURLItem)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});

context("when the app contain services and ingresses", () => {
  it("should show the table with available svcs and ingresses", () => {
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
    const ingress = {
      metadata: {
        name: "foo",
      },
      spec: {
        rules: [
          {
            host: "foo.bar",
            http: {
              paths: [{ path: "/ready" }],
            },
          },
        ],
      } as IIngressSpec,
    };
    const ingresses = {};
    ingresses[ingress.metadata.name] = ingress;
    const wrapper = shallow(<AccessURLTable services={services} ingresses={ingresses} />);
    expect(wrapper.find(AccessURLItem)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});
