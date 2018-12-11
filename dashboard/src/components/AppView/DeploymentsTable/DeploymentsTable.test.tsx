import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../../shared/specs";

import { IResource } from "shared/types";
import LoadingWrapper from "../../../components/LoadingWrapper";
import DeploymentItem from "./DeploymentItem";
import DeploymentTable from "./DeploymentsTable";

context("when fetching deployments", () => {
  itBehavesLike("aLoadingComponent", {
    component: DeploymentTable,
    props: {
      deployments: [{ isFetching: true }],
    },
  });
});

it("renders a message if there are no deployments", () => {
  const wrapper = shallow(<DeploymentTable deployments={[]} />);
  expect(
    wrapper
      .find(LoadingWrapper)
      .shallow()
      .find(DeploymentItem),
  ).not.toExist();
  expect(
    wrapper
      .find(LoadingWrapper)
      .shallow()
      .text(),
  ).toContain("The current application does not contain any deployment");
});

it("renders a deployment ready", () => {
  const deployments = [
    {
      isFetching: false,
      item: {
        kind: "Deployment",
        metadata: {
          name: "foo",
        },
        status: {},
      } as IResource,
    },
  ];
  const wrapper = shallow(<DeploymentTable deployments={deployments} />);
  expect(wrapper).toMatchSnapshot();
  expect(wrapper.find(DeploymentItem).key()).toContain("foo");
});

it("renders two deployments", () => {
  const deployments = [
    {
      isFetching: false,
      item: {
        kind: "Deployment",
        metadata: {
          name: "foo",
        },
        status: {},
      } as IResource,
    },
    {
      isFetching: false,
      item: {
        kind: "Deployment",
        metadata: {
          name: "bar",
        },
        status: {},
      } as IResource,
    },
  ];
  const wrapper = shallow(<DeploymentTable deployments={deployments} />);
  expect(wrapper.find(DeploymentItem).length).toBe(2);
  expect(
    wrapper
      .find(DeploymentItem)
      .at(0)
      .prop("deployment"),
  ).toBe(deployments[0].item);
  expect(
    wrapper
      .find(DeploymentItem)
      .at(1)
      .prop("deployment"),
  ).toBe(deployments[1].item);
});
