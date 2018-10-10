import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { IServiceBroker } from "shared/ServiceCatalog";
import { IServiceInstance } from "shared/ServiceInstance";
import InstanceListView from ".";
import itBehavesLike from "../../shared/specs";
import { ForbiddenError } from "../../shared/types";
import {
  ErrorSelector,
  MessageAlert,
  ServiceBrokersNotFoundAlert,
  ServiceCatalogNotInstalledAlert,
} from "../ErrorAlert";
import { InstanceCardList } from "./InstanceCardList";

const defaultProps = {
  brokers: { isFetching: false, list: [] },
  classes: { isFetching: false, list: [] },
  error: undefined,
  filter: "",
  getBrokers: jest.fn(),
  getClasses: jest.fn(),
  getInstances: jest.fn(),
  checkCatalogInstalled: jest.fn(),
  instances: { isFetching: false, list: [] },
  pushSearchFilter: jest.fn(),
  isServiceCatalogInstalled: true,
  namespace: "default",
};

it("renders the warning for alpha feature", () => {
  const wrapper = shallow(<InstanceListView {...defaultProps} />);
  expect(
    wrapper
      .find(MessageAlert)
      .children()
      .text(),
  ).toContain("Service Catalog integration is under heavy development");
});

context("while fetching components", () => {
  const props = { ...defaultProps, classes: { isFetching: true, list: [] } };

  itBehavesLike("aLoadingComponent", { component: InstanceListView, props });

  it("matches the snapshot", () => {
    const wrapper = shallow(<InstanceListView {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<InstanceListView {...props} />);
    expect(wrapper.find("h1").text()).toContain("Service Instances");
  });
});

context("when all the components are loaded", () => {
  it("shows a warning to install the Service Catalog if it's not installed", () => {
    const wrapper = shallow(
      <InstanceListView {...defaultProps} isServiceCatalogInstalled={false} />,
    );
    expect(wrapper.find(ServiceCatalogNotInstalledAlert)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("shows a warning to install no service broker is installed", () => {
    const wrapper = shallow(<InstanceListView {...defaultProps} />);
    expect(wrapper.find(ServiceBrokersNotFoundAlert)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("shows a forbiden error if it exists", () => {
    const wrapper = shallow(<InstanceListView {...defaultProps} error={new ForbiddenError()} />);
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("shows information of how to deploy a service instance if there are none of them", () => {
    const wrapper = shallow(
      <InstanceListView
        {...defaultProps}
        brokers={{ isFetching: false, list: [{} as IServiceBroker] }}
      />,
    );
    const message = wrapper.find(MessageAlert).filterWhere(m => {
      const header = m.prop("header");
      return (
        !!header && header === "Provision External Services from the Kubernetes Service Catalog"
      );
    });
    expect(message).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("shows a card list with the different instances if they exists", () => {
    const wrapper = shallow(
      <InstanceListView
        {...defaultProps}
        brokers={{ isFetching: false, list: [{} as IServiceBroker] }}
        instances={{
          isFetching: false,
          list: [
            {
              metadata: {
                name: "foo",
              },
            } as IServiceInstance,
          ],
        }}
      />,
    );
    expect(wrapper.find(InstanceCardList)).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});
