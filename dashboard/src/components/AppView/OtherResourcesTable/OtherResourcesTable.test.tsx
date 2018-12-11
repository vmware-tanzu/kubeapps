import { shallow } from "enzyme";
import * as React from "react";
import { IResource } from "../../../shared/types";
import OtherResourcesTable from "./OtherResourcesTable";

const defaultProps = {
  deployments: {},
  otherResources: {},
  services: {},
};

it("renders other resources details", () => {
  const otherResources = [
    {
      kind: "Secret",
      metadata: {
        name: "foo",
      },
      status: {},
    } as IResource,
    {
      kind: "PersistentVolume",
      metadata: {
        name: "foo",
      },
      status: {},
    } as IResource,
  ];
  const wrapper = shallow(
    <OtherResourcesTable {...defaultProps} otherResources={otherResources} />,
  );
  expect(wrapper).toMatchSnapshot();
});
