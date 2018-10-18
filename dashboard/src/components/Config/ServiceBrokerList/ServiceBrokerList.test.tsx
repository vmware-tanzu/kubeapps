import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import ServiceBrokerList from ".";
import { ErrorSelector, ServiceBrokersNotFoundAlert } from "../../../components/ErrorAlert";
import { IServiceBroker } from "../../../shared/ServiceCatalog";
import itBehavesLike from "../../../shared/specs";
import { ForbiddenError } from "../../../shared/types";

const defaultProps = { brokers: { isFetching: false, list: [] }, sync: jest.fn(), errors: {} };
const broker = {
  metadata: { name: "wall-street", uid: "1" },
  spec: {
    url: "https://foo.bar",
  },
  status: {
    lastCatalogRetrievalTime: "today",
  },
} as IServiceBroker;

context("while fetching brokers", () => {
  const props = { ...defaultProps, brokers: { isFetching: true, list: [] } };

  itBehavesLike("aLoadingComponent", { component: ServiceBrokerList, props });

  it("matches the snapshot", () => {
    const wrapper = shallow(<ServiceBrokerList {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<ServiceBrokerList {...props} />);
    expect(wrapper.find("h1").text()).toContain("Service Brokers");
  });
});

context("when all the brokers are loaded", () => {
  it("shows a warning to install no service broker is installed", () => {
    const wrapper = mount(<ServiceBrokerList {...defaultProps} />);
    expect(wrapper.find(ServiceBrokersNotFoundAlert)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("shows a forbiden (fetch) error if it exists", () => {
    const wrapper = shallow(
      <ServiceBrokerList {...defaultProps} errors={{ fetch: new ForbiddenError() }} />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("shows a forbiden (resync) error if it exists", () => {
    const wrapper = mount(
      <ServiceBrokerList
        {...defaultProps}
        brokers={{ isFetching: false, list: [broker] }}
        errors={{ update: new ForbiddenError() }}
      />,
    );
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("when there are no errors, renders the broker list", () => {
    const wrapper = shallow(
      <ServiceBrokerList {...defaultProps} brokers={{ list: [broker], isFetching: false }} />,
    );
    expect(wrapper).toMatchSnapshot();
  });
});
