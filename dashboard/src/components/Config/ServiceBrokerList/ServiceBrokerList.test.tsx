import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import ServiceBrokerList from ".";
import {
  ErrorSelector,
  MessageAlert,
  ServiceBrokersNotFoundAlert,
  ServiceCatalogNotInstalledAlert,
} from "../../../components/ErrorAlert";
import { IServiceBroker } from "../../../shared/ServiceCatalog";
import itBehavesLike from "../../../shared/specs";
import { ForbiddenError } from "../../../shared/types";

let defaultProps = {
  getBrokers: jest.fn(),
  brokers: { isFetching: false, list: [] },
  sync: jest.fn(),
  errors: {},
  checkCatalogInstalled: jest.fn(),
  isInstalled: true,
  cluster: "default",
  kubeappsCluster: "default",
};

beforeEach(() => {
  // Restart mock stats
  defaultProps = {
    getBrokers: jest.fn(),
    brokers: { isFetching: false, list: [] },
    sync: jest.fn(),
    errors: {},
    checkCatalogInstalled: jest.fn(),
    isInstalled: true,
    cluster: "default",
    kubeappsCluster: "default",
  };
});

const broker = {
  metadata: { name: "wall-street", uid: "1" },
  spec: {
    url: "https://foo.bar",
  },
  status: {
    lastCatalogRetrievalTime: "today",
  },
} as IServiceBroker;

context("if the service broker is not installed", () => {
  it("shows a warning message", () => {
    const props = { ...defaultProps, isInstalled: false };
    const wrapper = shallow(<ServiceBrokerList {...props} />);
    expect(wrapper.find(ServiceCatalogNotInstalledAlert)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});

context("if the service brokers are accessed on an additional cluster", () => {
  it("shows an alert with info", () => {
    const props = { ...defaultProps, cluster: "other-cluster" };
    const wrapper = shallow(<ServiceBrokerList {...props} />);
    const msgAlert = wrapper.find(MessageAlert);
    expect(msgAlert).toExist();
    expect(msgAlert.prop("header")).toEqual(
      "Service brokers can be created on the cluster on which Kubeapps is installed only",
    );
  });

  it("does not show an alert with info on the default cluster", () => {
    const wrapper = shallow(<ServiceBrokerList {...defaultProps} />);
    const msgAlert = wrapper.find(MessageAlert);
    expect(msgAlert).not.toExist();
  });
});

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
