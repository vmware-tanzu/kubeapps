import { shallow } from "enzyme";
import * as React from "react";

import { MessageAlert } from "../ErrorAlert";
import { InstanceListView } from "./index";

it("renders the warning for alpha feature", () => {
  const wrapper = shallow(
    <InstanceListView
      brokers={[]}
      classes={[]}
      error={undefined}
      filter={""}
      getCatalog={jest.fn()}
      checkCatalogInstalled={jest.fn()}
      instances={[]}
      plans={[]}
      pushSearchFilter={jest.fn()}
      isInstalled={true}
      namespace="default"
    />,
  );
  expect(
    wrapper
      .find(MessageAlert)
      .children()
      .text(),
  ).toContain("Service Catalog integration is under heavy development");
});
