import { shallow } from "enzyme";
import * as React from "react";
import * as ReactTooltip from "react-tooltip";

import { IKubeItem, IResource } from "shared/types";
import ApplicationStatus from "./ApplicationStatus";

const defaultProps = {
  watchWorkloads: jest.fn(),
  closeWatches: jest.fn(),
  deployments: [],
  statefulsets: [],
  daemonsets: [],
};

const consoleError = global.console.error;
beforeEach(() => {
  // Mute console.error since we are getting a lot of error for rendering the PieChart component
  // more info here: https://github.com/toomuchdesign/react-minimal-pie-chart/issues/131
  global.console.error = jest.fn();
});
afterEach(() => {
  jest.resetAllMocks();
  global.console.error = consoleError;
});

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
    totalPods: number;
    readyPods: number;
  }> = [
    {
      title: "shows a deployed status if there are no resources",
      deployments: [],
      statefulsets: [],
      daemonsets: [],
      deployed: true,
      totalPods: 0,
      readyPods: 0,
    },
    {
      title: "shows a deploying status if there is a non deployed deployment",
      deployments: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
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
      totalPods: 1,
      readyPods: 0,
    },
    {
      title: "shows a deploying status if there is a non deployed statefulset",
      statefulsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
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
      totalPods: 1,
      readyPods: 0,
    },
    {
      title: "shows a deploying status if there is a non deployed daemonset",
      daemonsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
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
      totalPods: 1,
      readyPods: 0,
    },
    {
      title: "shows a deployed status if it has a daemonset, deployment and statefulset deployed",
      daemonsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo" },
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
            metadata: { name: "foo" },
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
            metadata: { name: "foo" },
            status: {
              replicas: 1,
              readyReplicas: 1,
            },
          } as IResource,
        },
      ],
      deployed: true,
      totalPods: 3,
      readyPods: 3,
    },
    {
      title:
        "shows a deploying status if it has a daemonset, deployment (deployed) and statefulset (not deployed)",
      daemonsets: [
        {
          isFetching: false,
          item: {
            metadata: { name: "foo-ds" },
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
            metadata: { name: "foo-dp" },
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
            metadata: { name: "foo-ss" },
            status: {
              replicas: 1,
              readyReplicas: 0,
            },
          } as IResource,
        },
      ],
      deployed: true,
      totalPods: 3,
      readyPods: 2,
    },
  ];
  tests.forEach(t => {
    it(t.title, () => {
      const wrapper = shallow(<ApplicationStatus {...defaultProps} />);
      wrapper.setProps({
        deployments: t.deployments,
        statefulsets: t.statefulsets,
        daemonsets: t.daemonsets,
      });
      expect(wrapper.text()).toContain(t.deployed ? "Ready" : "Not Ready");
      expect(wrapper.state()).toMatchObject({ totalPods: t.totalPods, readyPods: t.readyPods });
      // Check tooltip text
      const tooltipText = wrapper
        .find(ReactTooltip)
        .dive()
        .text();
      t.deployments.forEach(d =>
        expect(tooltipText).toContain(
          `${d.item?.status.availableReplicas}/${d.item?.status.replicas}${d.item?.metadata.name}`,
        ),
      );
      t.statefulsets.forEach(d =>
        expect(tooltipText).toContain(
          `${d.item?.status.readyReplicas}/${d.item?.status.replicas}${d.item?.metadata.name}`,
        ),
      );
      t.daemonsets.forEach(d =>
        expect(tooltipText).toContain(
          `${d.item?.status.numberReady}/${d.item?.status.currentNumberScheduled}${d.item?.metadata.name}`,
        ),
      );
    });
  });
});
