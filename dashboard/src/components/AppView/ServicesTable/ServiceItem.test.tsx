import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";
import { IKubeItem, IResource } from "../../../shared/types";
import ServiceItem from "./ServiceItem";

const kubeItem: IKubeItem<IResource> = {
  isFetching: false,
};

context("when fetching services", () => {
  [undefined, { isFetching: true }].forEach(service => {
    itBehavesLike("aLoadingComponent", {
      component: ServiceItem,
      props: {
        service,
        getService: jest.fn(),
      },
    });
    it("displays the name of the Service", () => {
      const wrapper = shallow(<ServiceItem service={service} name="foo" getService={jest.fn()} />);
      expect(wrapper.text()).toContain("foo");
    });
  });
});

context("when there is an error fetching the Service", () => {
  const service = {
    error: new Error('services "foo" not found'),
    isFetching: false,
  };
  const wrapper = shallow(<ServiceItem service={service} name="foo" getService={jest.fn()} />);

  it("diplays the Service name in the first column", () => {
    expect(
      wrapper
        .find("td")
        .first()
        .text(),
    ).toEqual("foo");
  });

  it("displays the error message in the second column", () => {
    expect(
      wrapper
        .find("td")
        .at(1)
        .text(),
    ).toContain('Error: services "foo" not found');
  });
});

context("when there is a valid Service", () => {
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
      <ServiceItem service={kubeItem} name={service.metadata.name} getService={jest.fn()} />,
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
      <ServiceItem service={kubeItem} name={service.metadata.name} getService={jest.fn()} />,
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
      <ServiceItem service={kubeItem} name={service.metadata.name} getService={jest.fn()} />,
    );
    expect(wrapper).toMatchSnapshot();
  });
});
