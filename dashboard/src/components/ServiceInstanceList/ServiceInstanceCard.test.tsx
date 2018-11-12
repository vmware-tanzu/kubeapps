import { shallow } from "enzyme";
import * as React from "react";

import InfoCard from "../InfoCard";
import ServiceInstanceCard from "./ServiceInstanceCard";

it("should render a Card", () => {
  const wrapper = shallow(
    <ServiceInstanceCard
      key="1"
      name="foo"
      namespace="foobar"
      link="/a/link/somewhere"
      icon="an-icon.png"
      serviceClassName="database"
      servicePlanName="large"
      statusReason="Provisioned"
    />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("should avoid the status tag if it's not defined", () => {
  const wrapper = shallow(
    <ServiceInstanceCard
      key="1"
      name="foo"
      namespace="foobar"
      icon="an-icon.png"
      link="/a/link/somewhere"
      serviceClassName="database"
      servicePlanName="large"
      statusReason={undefined}
    />,
  );
  expect(wrapper.find(".ChartListItem__content__info_other")).not.toExist();
});

describe("uses a different class and tag name depending on the status", () => {
  [
    {
      description: "Uses Provisioned when the status is alike",
      status: "SuccessfullyProvisioned",
      expected: "Provisioned",
    },
    {
      description: "Uses Failed when the status is alike",
      status: "CreationFailed",
      expected: "Failed",
    },
    {
      description: "Uses the raw status when it's unknown",
      status: "StillProvisioning",
      expected: "StillProvisioning",
    },
  ].forEach(test => {
    it(test.description, () => {
      const wrapper = shallow(
        <ServiceInstanceCard
          key="1"
          name="foo"
          namespace="foobar"
          link="/a/link/somewhere"
          icon="an-icon.png"
          serviceClassName="database"
          servicePlanName="large"
          statusReason={test.status}
        />,
      );
      const card = wrapper.find(InfoCard);
      expect(card).toExist();
      expect(card.props()).toMatchObject({ tag2Class: test.expected, tag2Content: test.expected });
    });
  });
});
