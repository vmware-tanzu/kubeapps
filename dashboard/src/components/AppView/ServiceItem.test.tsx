import { shallow } from "enzyme";
import * as React from "react";

import { IResource } from "shared/types";
import ServiceItem from "./ServiceItem";

it("renders a simple view without IP", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "ClusterIP",
    },
  } as IResource;
  const wrapper = shallow(<ServiceItem service={service} extended={false} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(".ServiceItem__url").text()).toBe("None");
});

it("renders a simple view with IP", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "LoadBalancer",
      ports: [{ port: 80 }],
    },
    status: {
      loadBalancer: {
        ingress: [{ ip: "1.2.3.4" }],
      },
    },
  } as IResource;
  const wrapper = shallow(<ServiceItem service={service} extended={false} />);
  expect(wrapper.find(".ServiceItem__url").text()).toBe("http://1.2.3.4");
});

it("renders an extended view with IP", () => {
  const service = {
    metadata: {
      name: "foo",
    },
    spec: {
      type: "LoadBalancer",
      ports: [{ port: 80, protocol: "http" }],
    },
    status: {
      loadBalancer: {
        ingress: [{ ip: "1.2.3.4" }],
      },
    },
  } as IResource;
  const wrapper = shallow(<ServiceItem service={service} extended={true} />);
  expect(wrapper).toMatchSnapshot();
});
