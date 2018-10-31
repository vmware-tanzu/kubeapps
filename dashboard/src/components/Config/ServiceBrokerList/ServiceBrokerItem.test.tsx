import { shallow } from "enzyme";
import * as React from "react";

import { IServiceBroker } from "../../../shared/ServiceCatalog";
import ServiceBrokerItem from "./ServiceBrokerItem";

it("renders a card with the service broker", () => {
  const broker = {
    metadata: { name: "wall-street" },
    spec: {
      url: "https://foo.bar",
    },
    status: {
      lastCatalogRetrievalTime: "today",
    },
  } as IServiceBroker;
  const wrapper = shallow(<ServiceBrokerItem broker={broker} sync={jest.fn()} />);
  expect(wrapper).toMatchSnapshot();
});
