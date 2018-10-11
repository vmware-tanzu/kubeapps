import { shallow } from "enzyme";
import * as React from "react";

import { MessageAlert } from "../ErrorAlert";
import ServiceInstanceList, { IServiceInstanceListProps } from "./ServiceInstanceList";

it("renders the warning for alpha feature", () => {
  const wrapper = shallow(
    <ServiceInstanceList
      {...{} as IServiceInstanceListProps}
      brokers={[]}
      classes={{ isFetching: false, list: [] }}
      getCatalog={jest.fn()}
      checkCatalogInstalled={jest.fn()}
      instances={[]}
      plans={[]}
      pushSearchFilter={jest.fn()}
      isInstalled={true}
    />,
  );
  expect(
    wrapper
      .find(MessageAlert)
      .children()
      .text(),
  ).toContain("Service Catalog integration is under heavy development");
});
