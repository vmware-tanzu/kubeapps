import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { IResource } from "../../../../../shared/types";
import ServiceItem from "./ServiceItem";

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
    const wrapper = shallow(<ServiceItem resource={service} />);
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
    const wrapper = shallow(<ServiceItem resource={service} />);
    expect(wrapper.text()).toContain("1.2.3.4");
  });

  it("renders a view with IP", () => {
    const service = {
      metadata: { name: "foo" },
      spec: { ports: [{ port: 80 }], type: "LoadBalancer" },
      status: { loadBalancer: { ingress: [{ ip: "1.2.3.4" }] } },
    } as IResource;
    const wrapper = shallow(<ServiceItem resource={service} />);
    expect(wrapper).toMatchSnapshot();
  });
});
