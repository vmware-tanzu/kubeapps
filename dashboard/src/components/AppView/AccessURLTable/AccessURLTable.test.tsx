import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";

import ResourceRef from "shared/ResourceRef";
import { IIngressSpec, IResource, IServiceSpec, IServiceStatus } from "../../../shared/types";
import AccessURLItem from "./AccessURLItem";
import AccessURLTable from "./AccessURLTable";

const defaultProps = {
  services: [],
  ingresses: [],
  ingressRefs: [],
  getResource: jest.fn(),
};

describe("when receiving ingresses", () => {
  it("fetches ingresses at mount time", () => {
    const ingress = { name: "ing" } as ResourceRef;
    const mock = jest.fn();
    shallow(<AccessURLTable {...defaultProps} ingressRefs={[ingress]} getResource={mock} />);
    expect(mock).toHaveBeenCalledWith(ingress);
  });

  it("fetches when new ingress refs received", () => {
    const ingress = { name: "ing" } as ResourceRef;
    const mock = jest.fn();
    const wrapper = shallow(<AccessURLTable {...defaultProps} getResource={mock} />);
    wrapper.setProps({ ingressRefs: [ingress] });
    expect(mock).toHaveBeenCalledWith(ingress);
  });
});

context("when fetching ingresses or services", () => {
  itBehavesLike("aLoadingComponent", {
    component: AccessURLTable,
    props: {
      ...defaultProps,
      ingresses: [{ isFetching: true }],
    },
  });
  itBehavesLike("aLoadingComponent", {
    component: AccessURLTable,
    props: {
      ...defaultProps,
      services: [{ isFetching: true }],
    },
  });
});

it("doesn't render anything if the application has no URL", () => {
  const wrapper = shallow(<AccessURLTable {...defaultProps} />);
  expect(wrapper.find("table")).not.toExist();
});

context("when the app contains services", () => {
  it("should omit the Service Table if there are no public services", () => {
    const service = {
      kind: "Service",
      metadata: {
        name: "foo",
        selfLink: "/services/foo",
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
    const wrapper = shallow(<AccessURLTable {...defaultProps} services={services} />);
    expect(wrapper.text()).toContain("The current application does not expose a public URL");
  });

  it("should show the table if any service is a LoadBalancer", () => {
    const service = {
      kind: "Service",
      metadata: {
        name: "foo",
        selfLink: "/services/foo",
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
    const wrapper = shallow(<AccessURLTable {...defaultProps} services={services} />);
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
        selfLink: "/ingresses/foo",
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
    const wrapper = shallow(<AccessURLTable {...defaultProps} ingresses={ingresses} />);
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
        selfLink: "/services/foo",
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
        selfLink: "/ingresses/foo",
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
      <AccessURLTable {...defaultProps} services={services} ingresses={ingresses} />,
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
      <AccessURLTable {...defaultProps} services={services} ingresses={ingresses} />,
    );

    // The Service error is not shown, as it is filtered out because without the
    // resource we can't determine whether it is a public LoadBalancer Service
    // or not. The Service error will be shown in the Services table anyway.
    expect(wrapper.find(AccessURLItem)).not.toExist();
    expect(wrapper.find("table").text()).toContain("could not find Ingress");

    expect(wrapper).toMatchSnapshot();
  });
});
