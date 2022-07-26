// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import { app } from "shared/url";
import InfoCard from "../InfoCard/InfoCard";
import CustomResourceListItem from "./CustomResourceListItem";

const defaultProps = {
  cluster: "default",
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
    icon: "placeholder.svg",
    link: app.operatorInstances.view(
      defaultProps.cluster,
      resource.metadata.namespace,
      csv.metadata.name,
      crd.name,
      resource.metadata.name,
    ),
    title: resource.metadata.name,
  });
});
