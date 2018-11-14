import { shallow } from "enzyme";
import * as React from "react";

import { IServiceInstance } from "../../shared/ServiceInstance";
import ServiceInstanceStatus from "./ServiceInstanceStatus";

const baseInstance = {
  status: {
    conditions: [
      {
        lastTransitionTime: "1",
        type: "a type",
        status: "good",
        reason: "none",
        message: "everything okay here",
      },
    ],
  },
} as IServiceInstance;

const testCases = [
  { name: "provisioning", reason: "Provisioning", want: "Provisioning" },
  { name: "provisioned successfully", reason: "ProvisionedSuccessfully", want: "Provisioned" },
  { name: "failed", reason: "Failed provisioning", want: "Failed" },
  { name: "error", reason: "Error provisioning", want: "Failed" },
  { name: "deprovisioning", reason: "Deprovisioning", want: "Deprovisioning" },
  { name: "unknown", reason: "Random other reason", want: "Unknown" },
];

testCases.forEach(t => {
  it(`renders correct statuses when ${t.name}`, () => {
    const instance = {
      ...baseInstance,
      status: { ...baseInstance.status, conditions: [{ reason: t.reason }] },
    } as IServiceInstance;
    const wrapper = shallow(<ServiceInstanceStatus instance={instance} />);
    expect(wrapper.text()).toContain(t.want);
    expect(wrapper).toMatchSnapshot();
  });
});

it("displays unknown status when conditions is empty", () => {
  const instance = { ...baseInstance, status: { conditions: [] } } as IServiceInstance;
  const wrapper = shallow(<ServiceInstanceStatus instance={instance} />);
  expect(wrapper.text()).toContain("Unknown");
  expect(wrapper).toMatchSnapshot();
});
