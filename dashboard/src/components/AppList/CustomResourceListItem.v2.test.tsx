import { shallow } from "enzyme";
import * as React from "react";
import { app } from "shared/url";
import InfoCard from "../InfoCard/InfoCard.v2";
import CustomResourceListItem from "./CustomResourceListItem.v2";

const defaultProps = {
  resource: {
    kind: "Something",
    metadata: {
      name: "foo",
      namespace: "default",
    },
  },
  csv: {
    metadata: {
      name: "bar",
      namespace: "default",
    },
    spec: {
      version: "1.0.0",
      customresourcedefinitions: {
        owned: [
          {
            kind: "Something",
            name: "something",
            description: "it's something",
          },
        ],
      },
    },
  },
} as any;

const { resource, csv } = defaultProps;
const crd = defaultProps.csv.spec.customresourcedefinitions.owned[0];

it("renders an cr item", () => {
  const wrapper = shallow(<CustomResourceListItem {...defaultProps} />);
  const card = wrapper.find(InfoCard);
  expect(card.props()).toMatchObject({
    description: crd.description,
    icon: "placeholder.png",
    link: app.operatorInstances.view(
      resource.metadata.namespace,
      csv.metadata.name,
      crd.name,
      resource.metadata.name,
    ),
    tag1Content: "bar",
    title: resource.metadata.name,
  });
});
