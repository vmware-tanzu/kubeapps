import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";

import LoadingWrapper from "../../../components/LoadingWrapper";
import { IIngressSpec, IResource, IServiceSpec, IServiceStatus } from "../../../shared/types";
import AccessURLItem from "./AccessURLItem";
import AccessURLTable from "./AccessURLTable";

context("when fetching ingresses or services", () => {
  itBehavesLike("aLoadingComponent", {
    component: AccessURLTable,
    props: {
      ingresses: [{ isFetching: true }],
      services: [],
    },
  });
  itBehavesLike("aLoadingComponent", {
    component: AccessURLTable,
    props: {
      ingresses: [],
      services: [{ isFetching: true }],
    },
  });
});

it("renders a message if there are no services or ingresses", () => {
  const wrapper = shallow(<AccessURLTable services={[]} ingresses={[]} />);
  expect(
    wrapper
      .find(LoadingWrapper)
      .shallow()
      .find(AccessURLItem),
  ).not.toExist();
  expect(
    wrapper
      .find(LoadingWrapper)
      .shallow()
      .text(),
  ).toContain("The current application does not expose a public URL");
});

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
    const services = [{ isFetching: false, item: service }];
    const wrapper = shallow(<AccessURLTable services={services} ingresses={[]} />);
    expect(
      wrapper
        .find(LoadingWrapper)
        .shallow()
        .text(),
    ).toContain("The current application does not expose a public URL");
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
    const services = [{ isFetching: false, item: service }];
    const wrapper = shallow(<AccessURLTable services={services} ingresses={[]} />);
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
    } as IResource;
    const ingresses = [{ isFetching: false, item: ingress }];
    const wrapper = shallow(<AccessURLTable services={[]} ingresses={ingresses} />);
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
    const services = [{ isFetching: false, item: service }];
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
    } as IResource;
    const ingresses = [{ isFetching: false, item: ingress }];
    const wrapper = shallow(<AccessURLTable services={services} ingresses={ingresses} />);
    expect(wrapper.find(AccessURLItem)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});
