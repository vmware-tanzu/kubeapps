import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { IClusterServiceClass } from "../../shared/ClusterServiceClass";
import itBehavesLike from "../../shared/specs";
import { CardGrid } from "../Card";
import { MessageAlert } from "../ErrorAlert";
import ServiceClassList from "./ServiceClassList";

const defaultProps = {
  error: undefined,
  classes: {
    list: [],
    isFetching: false,
  },
  getClasses: jest.fn(),
};

it("should show a warning message if there are no classes available", () => {
  const wrapper = shallow(<ServiceClassList {...defaultProps} />);
  const alert = wrapper.find(MessageAlert);
  expect(alert).toExist();
  expect(alert.html()).toContain("Service Classes not found");
  expect(wrapper).toMatchSnapshot();
});

context("while fetching classes", () => {
  const props = { ...defaultProps, classes: { isFetching: true, list: [] } };

  itBehavesLike("aLoadingComponent", { component: ServiceClassList, props });

  it("matches the snapshot", () => {
    const wrapper = shallow(<ServiceClassList {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<ServiceClassList {...props} />);
    expect(wrapper.find("h1").text()).toContain("Classes");
  });
});

context("when there are classes available", () => {
  it("should show a Card item per class", () => {
    const class1 = {
      metadata: {
        uid: "1",
      },
      spec: {
        clusterServiceBrokerName: "azure",
        externalName: "foo",
        description: "this is a service!",
        externalMetadata: {
          imageUrl: "http://foo-image",
        },
      },
    } as IClusterServiceClass;
    const class2 = {
      metadata: {
        uid: "2",
      },
      spec: {
        clusterServiceBrokerName: "gcp",
        externalName: "bar",
        description: "this is a service!",
        externalMetadata: {
          imageUrl: "http://bar-image",
        },
      },
    } as IClusterServiceClass;
    const wrapper = shallow(
      <ServiceClassList
        {...defaultProps}
        classes={{ isFetching: false, list: [class1, class2] }}
      />,
    );
    const cardGrid = wrapper.find(CardGrid);
    expect(cardGrid).toExist();
    expect(cardGrid.children().length).toBe(2);
    expect(wrapper).toMatchSnapshot();
  });
});
