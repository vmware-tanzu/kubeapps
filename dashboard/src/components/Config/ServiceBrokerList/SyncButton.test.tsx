import { shallow } from "enzyme";
import * as React from "react";

import { IServiceBroker } from "../../../shared/ServiceCatalog";
import SyncButton from "./SyncButton";

const broker = {
  metadata: { name: "wall-street", uid: "1" },
  spec: {
    url: "https://foo.bar",
  },
  status: {
    lastCatalogRetrievalTime: "today",
  },
} as IServiceBroker;

it("should render the sync button", () => {
  const wrapper = shallow(<SyncButton broker={broker} sync={jest.fn()} />);
  expect(wrapper).toMatchSnapshot();
});

it("should execute the sync function with the given broker", done => {
  const sync = jest.fn(() => new Promise(s => s()));
  const wrapper = shallow(<SyncButton broker={broker} sync={sync} />);
  wrapper.simulate("click");
  expect(sync).toHaveBeenLastCalledWith(broker);
  expect((wrapper.state() as any).isSyncing).toBe(true);
  // Wait for sync to finish
  setTimeout(() => {
    expect((wrapper.state() as any).isSyncing).toBe(false);
    done();
  }, 1);
});
