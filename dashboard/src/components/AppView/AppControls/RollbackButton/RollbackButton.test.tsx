import { mount, shallow, ShallowWrapper } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import * as ReactModal from "react-modal";
import { IAppRepository, IChartVersion, IRelease } from "shared/types";
import RollbackButton from ".";
import SelectRepoForm from "../../../../components/SelectRepoForm";
import RollbackDialog from "./RollbackDialog";

const defaultProps = {
  app: {
    chart: {
      metadata: {
        name: "foo",
        version: "1.0.0",
      },
    },
  } as IRelease,
  rollbackApp: jest.fn(),
  getChartVersion: jest.fn(),
  loading: false,
  repos: [],
  repo: {} as IAppRepository,
  kubeappsNamespace: "kubeapps",
  fetchRepositories: jest.fn(),
  checkChart: jest.fn(),
};

it("stores the chart name and version in the state", () => {
  const wrapper = shallow(<RollbackButton {...defaultProps} />);

  expect(wrapper.state()).toMatchObject({ chartName: "foo", chartVersion: "1.0.0" });
});

function openModal(wrapper: ShallowWrapper) {
  ReactModal.setAppElement(document.createElement("div"));
  wrapper.setState({ modalIsOpen: true });
  wrapper.update();
}

context("when opening the Modal", () => {
  context("when there is no info about the chart", () => {
    it("should render the SelectRepoForm", () => {
      const wrapper = shallow(<RollbackButton {...defaultProps} chartVersion={undefined} />);
      openModal(wrapper);

      expect(wrapper.find(SelectRepoForm)).toExist();
      expect(wrapper.find(RollbackDialog)).not.toExist();
    });

    it("should fetch the available repositories", () => {
      const fetchRepositories = jest.fn();
      const wrapper = shallow(
        <RollbackButton
          {...defaultProps}
          chartVersion={undefined}
          fetchRepositories={fetchRepositories}
        />,
      );
      const button = wrapper.find(".button");
      expect(button).toExist();
      button.simulate("click");
      expect(fetchRepositories).toBeCalled();
    });

    it("if the app has updateInfo, fetch the chart", () => {
      const getChartVersion = jest.fn();
      const app = {
        chart: {
          metadata: {
            name: "bar",
            version: "1.0.0",
          },
        },
        updateInfo: {
          repository: {
            name: "foo",
          },
        },
      } as IRelease;
      const wrapper = mount(
        <RollbackButton
          {...defaultProps}
          chartVersion={undefined}
          getChartVersion={getChartVersion}
          app={app}
        />,
      );

      const button = wrapper.find(".button");
      expect(button).toExist();
      button.simulate("click");
      expect(getChartVersion).toBeCalledWith("foo/bar", "1.0.0");
    });

    it("should retrieve a chart if there is no updateInfo through the SelectRepoForm", async () => {
      const checkChart = jest.fn().mockReturnValueOnce(true);
      const getChartVersion = jest.fn();
      const wrapper = shallow(
        <RollbackButton
          {...defaultProps}
          chartVersion={undefined}
          checkChart={checkChart}
          getChartVersion={getChartVersion}
        />,
      );
      openModal(wrapper);
      wrapper.setState({ chartVersion: "1.0.0" });

      const form = wrapper.find(SelectRepoForm);
      expect(form).toExist();
      const checkChartWithinForm = form.prop("checkChart") as (
        repo: string,
        chartName: string,
      ) => any;
      await checkChartWithinForm("foo", "bar");
      expect(checkChart).toBeCalled();
      expect(getChartVersion).toBeCalledWith("foo/bar", "1.0.0");
    });
  });

  context("when there is info about the chart", () => {
    it("should render the RollbackDialog", () => {
      const wrapper = shallow(
        <RollbackButton {...defaultProps} chartVersion={{} as IChartVersion} />,
      );
      openModal(wrapper);

      expect(wrapper.find(SelectRepoForm)).not.toExist();
      expect(wrapper.find(RollbackDialog)).toExist();
    });

    it("should perform the rollback", async () => {
      const rollbackApp = jest.fn();
      const chart = { id: "foo" } as IChartVersion;
      const app = {
        name: "bar",
        namespace: "default",
        config: { raw: "this: that" },
        chart: {
          metadata: {
            name: "foo",
            version: "1.0.0",
          },
        },
      } as IRelease;
      const wrapper = shallow(
        <RollbackButton
          {...defaultProps}
          rollbackApp={rollbackApp}
          chartVersion={chart}
          app={app}
        />,
      );
      openModal(wrapper);

      const dialog = wrapper.find(RollbackDialog);
      expect(dialog).toExist();
      const onConfirm = dialog.prop("onConfirm") as (revision: number) => Promise<any>;
      await onConfirm(1);
      expect(rollbackApp).toBeCalledWith(chart, app.name, 1, app.namespace, app.config!.raw);
    });
  });
});
