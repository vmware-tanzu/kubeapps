import { shallow } from "enzyme";
import * as React from "react";

import { IKubeItem, IResource } from "shared/types";
import ApplicationStatus from "./ApplicationStatus";

const defaultProps = {
  watchWorkloads: jest.fn(),
  closeWatches: jest.fn(),
  deployments: [],
  statefulsets: [],
  daemonsets: [],
};

describe("componentDidMount", () => {
  it("calls watchWorkloads", () => {
    const mock = jest.fn();
    shallow(<ApplicationStatus {...defaultProps} watchWorkloads={mock} />);
    expect(mock).toHaveBeenCalled();
  });
});

describe("componentWillUnmount", () => {
  it("calls closeWatches", () => {
    const mock = jest.fn();
    const wrapper = shallow(<ApplicationStatus {...defaultProps} closeWatches={mock} />);
    wrapper.unmount();
    expect(mock).toHaveBeenCalled();
  });
});

it("renders a loading status", () => {
  const deployments = [
    {
      isFetching: true,
    },
  ];
  const wrapper = shallow(<ApplicationStatus {...defaultProps} deployments={deployments} />);
  expect(wrapper.text()).toContain("Loading");
  expect(wrapper).toMatchSnapshot();
});

it("renders a deleting status", () => {
  const deployments = [
    {
      isFetching: false,
    },
  ];
  const wrapper = shallow(
    <ApplicationStatus {...defaultProps} deployments={deployments} info={{ deleted: {} }} />,
  );
  expect(wrapper.text()).toContain("Deleted");
  expect(wrapper).toMatchSnapshot();
});

it("renders a failed status", () => {
  const deployments = [
    {
      isFetching: false,
    },
  ];
  const wrapper = shallow(
    <ApplicationStatus
      {...defaultProps}
      deployments={deployments}
      info={{ status: { code: 4 } }}
    />,
  );
  expect(wrapper.text()).toContain("Failed");
  expect(wrapper).toMatchSnapshot();
});

describe("isFetching", () => {
  const tests: Array<{
    title: string;
    deployments: Array<IKubeItem<IResource>>;
    statefulsets: Array<IKubeItem<IResource>>;
    daemonsets: Array<IKubeItem<IResource>>;
    deployed: boolean;
  }> = [
    {
      title: "shows a deployed status if there are no resources",
      deployments: [],
      statefulsets: [],
      daemonsets: [],
      deployed: true,
    },
    {
      title: "shows a deploying status if there is a non deployed deployment",
      deployments: [
        {
          isFetching: false,
          item: {
            status: {
              replicas: 1,
              availableReplicas: 0,
            },
          } as IResource,
        },
      ],
      statefulsets: [],
      daemonsets: [],
      deployed: false,
    },
    {
      title: "shows a deploying status if there is a non deployed statefulset",
      statefulsets: [
        {
          isFetching: false,
          item: {
            status: {
              replicas: 1,
              readyReplicas: 0,
            },
          } as IResource,
        },
      ],
      deployments: [],
      daemonsets: [],
      deployed: false,
    },
    {
      title: "shows a deploying status if there is a non deployed daemonset",
      daemonsets: [
        {
          isFetching: false,
          item: {
            status: {
              currentNumberScheduled: 1,
              numberReady: 0,
            },
          } as IResource,
        },
      ],
      deployments: [],
      statefulsets: [],
      deployed: false,
    },
    {
      title: "shows a deployed status if it has a daemonset, deployment and statefulset deployed",
      daemonsets: [
        {
          isFetching: false,
          item: {
            status: {
              currentNumberScheduled: 1,
              numberReady: 1,
            },
          } as IResource,
        },
      ],
      deployments: [
        {
          isFetching: false,
          item: {
            status: {
              replicas: 1,
              availableReplicas: 1,
            },
          } as IResource,
        },
      ],
      statefulsets: [
        {
          isFetching: false,
          item: {
            status: {
              replicas: 1,
              readyReplicas: 1,
            },
          } as IResource,
        },
      ],
      deployed: true,
    },
  ];
  tests.forEach(t => {
    it(t.title, () => {
      const wrapper = shallow(
        <ApplicationStatus
          {...defaultProps}
          deployments={t.deployments}
          statefulsets={t.statefulsets}
          daemonsets={t.daemonsets}
        />,
      );
      expect(wrapper.text()).toContain(t.deployed ? "Ready" : "Not Ready");
    });
  });
});
