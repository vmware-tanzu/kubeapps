import { shallow } from "enzyme";
import * as React from "react";

import { IResource, IServiceSpec, IServiceStatus } from "shared/types";
import AccessURLItem from "./AccessURLItem";

describe("when the status is empty", () => {
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

  it("should show a Pending text", () => {
    const wrapper = shallow(<AccessURLItem service={service} />);
    expect(wrapper.text()).toContain("Pending");
    expect(wrapper).toMatchSnapshot();
  });

  it("should not include a link", () => {
    const wrapper = shallow(<AccessURLItem service={service} />);
    expect(wrapper.find(".ServiceItem__not-url")).toExist();
    expect(wrapper.find(".ServiceItem__url")).not.toExist();
    expect(wrapper.find("a")).not.toExist();
  });
});

describe("when the status is populated", () => {
  interface Itest {
    description: string;
    ports: any[];
    ingress: any[];
    expectedURLs: string[];
  }
  const tests: Itest[] = [
    {
      description: "it should show the IP and port if it's not known",
      ports: [{ port: 8080 }],
      ingress: [{ ip: "1.2.3.4" }],
      expectedURLs: ["http://1.2.3.4:8080"],
    },
    {
      description: "it should show the hostname and port if it's not known",
      ports: [{ port: 8080 }],
      ingress: [{ hostname: "1.2.3.4" }],
      expectedURLs: ["http://1.2.3.4:8080"],
    },
    {
      description: "it should show the IP and skip the port if it's known",
      ports: [{ port: 80 }],
      ingress: [{ ip: "1.2.3.4" }],
      expectedURLs: ["http://1.2.3.4"],
    },
    {
      description: "it should show the https URL if the port is 443",
      ports: [{ port: 443 }],
      ingress: [{ ip: "1.2.3.4" }],
      expectedURLs: ["https://1.2.3.4"],
    },
    {
      description: "it should show several URLs if there are multipe ports",
      ports: [{ port: 8080 }, { port: 8081 }],
      ingress: [{ ip: "1.2.3.4" }],
      expectedURLs: ["http://1.2.3.4:8080", "http://1.2.3.4:8081"],
    },
    {
      description: "it should show several URLs if there are ingress ports",
      ports: [{ port: 8080 }, { port: 8081 }],
      ingress: [{ ip: "1.2.3.4" }, { hostname: "foo.bar" }],
      expectedURLs: [
        "http://1.2.3.4:8080",
        "http://1.2.3.4:8081",
        "http://foo.bar:8080",
        "http://foo.bar:8081",
      ],
    },
  ];
  tests.forEach(test => {
    it(test.description, () => {
      const service = {
        metadata: {
          name: "foo",
        },
        spec: {
          type: "LoadBalancer",
          ports: test.ports,
        },
        status: {
          loadBalancer: {
            ingress: test.ingress,
          },
        } as IServiceStatus,
      } as IResource;
      const wrapper = shallow(<AccessURLItem service={service} />);
      test.expectedURLs.forEach(url => {
        expect(wrapper.find(".ServiceItem__not-url")).not.toExist();
        expect(wrapper.find(".ServiceItem__url")).toExist();
        expect(wrapper.text()).toContain(url);
      });
    });
  });
});
