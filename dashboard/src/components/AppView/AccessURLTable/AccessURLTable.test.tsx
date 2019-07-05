import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";

import LoadingWrapper from "../../../components/LoadingWrapper";
import { IIngressSpec, IResource, IServiceSpec, IServiceStatus } from "../../../shared/types";
import AccessURLItem from "./AccessURLItem";
import AccessURLTable from "./AccessURLTable";

describe("componentDidMount", () => {
  it("fetches ingresses", () => {
    const mock = jest.fn();
    shallow(<AccessURLTable services={[]} ingresses={[]} fetchIngresses={mock} />);
    expect(mock).toHaveBeenCalled();
  });
});

context("when fetching ingresses or services", () => {
  itBehavesLike("aLoadingComponent", {
    component: AccessURLTable,
    props: {
      ingresses: [{ isFetching: true }],
      services: [],
      fetchIngresses: jest.fn(),
      watchServices: jest.fn(),
      closeWatches: jest.fn(),
    },
  });
  itBehavesLike("aLoadingComponent", {
    component: AccessURLTable,
    props: {
      ingresses: [],
      services: [{ isFetching: true }],
      fetchIngresses: jest.fn(),
      watchServices: jest.fn(),
      closeWatches: jest.fn(),
    },
  });
});

it("renders a message if there are no services or ingresses", () => {
  const wrapper = shallow(
    <AccessURLTable services={[]} ingresses={[]} fetchIngresses={jest.fn()} />,
  );
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

context("when the app contains services", () => {
  it("should omit the Service Table if there are no public services", () => {
    const service = {
      kind: "Service",
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
    const wrapper = shallow(
      <AccessURLTable services={services} ingresses={[]} fetchIngresses={jest.fn()} />,
    );
    expect(
      wrapper
        .find(LoadingWrapper)
        .shallow()
        .text(),
    ).toContain("The current application does not expose a public URL");
  });

  it("should show the table if any service is a LoadBalancer", () => {
    const service = {
      kind: "Service",
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
    const wrapper = shallow(
      <AccessURLTable services={services} ingresses={[]} fetchIngresses={jest.fn()} />,
    );
    expect(wrapper.find(AccessURLItem)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});

context("when the app contains ingresses", () => {
  it("should show the table with available ingresses", () => {
    const ingress = {
      kind: "Ingress",
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
    const wrapper = shallow(
      <AccessURLTable services={[]} ingresses={ingresses} fetchIngresses={jest.fn()} />,
    );
    expect(wrapper.find(AccessURLItem)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});

context("when the app contains services and ingresses", () => {
  it("should show the table with available svcs and ingresses", () => {
    const service = {
      kind: "Service",
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
      kind: "Ingress",
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
    const wrapper = shallow(
      <AccessURLTable services={services} ingresses={ingresses} fetchIngresses={jest.fn()} />,
    );
    expect(wrapper.find(AccessURLItem)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});

context("when the app contains resources with errors", () => {
  it("displays the error", () => {
    const services = [{ isFetching: false, error: new Error("could not find Service") }];
    const ingresses = [{ isFetching: false, error: new Error("could not find Ingress") }];
    const wrapper = shallow(
      <AccessURLTable services={services} ingresses={ingresses} fetchIngresses={jest.fn()} />,
    );

    // The Service error is not shown, as it is filtered out because without the
    // resource we can't determine whether it is a public LoadBalancer Service
    // or not. The Service error will be shown in the Services table anyway.
    expect(wrapper.find(AccessURLItem)).not.toExist();
    expect(wrapper.find("table").text()).toContain("could not find Ingress");

    expect(wrapper).toMatchSnapshot();
  });
});
