import * as React from "react";
import { RouterAction } from "react-router-redux";

import { IServiceBinding } from "../../shared/ServiceBinding";
import { IChartState, IChartVersion } from "../../shared/types";

import DeploymentForm from "../DeploymentForm";

import "./AppNew.css";

interface IAppNewProps {
  bindings: IServiceBinding[];
  chartID: string;
  deployChart: (
    version: IChartVersion,
    releaseName: string,
    namespace: string,
    values?: string,
  ) => Promise<{}>;
  selected: IChartState["selected"];
  chartVersion: string;
  push: (location: string) => RouterAction;
  fetchChartVersions: (id: string) => Promise<{}>;
  getBindings: () => Promise<IServiceBinding[]>;
  getChartVersion: (id: string, chartVersion: string) => Promise<{}>;
  getChartValues: (id: string, chartVersion: string) => Promise<{}>;
  selectChartVersionAndGetFiles: (version: IChartVersion) => Promise<{}>;
}

interface IAppNewState {
  error?: string;
}

class AppNew extends React.Component<IAppNewProps, IAppNewState> {
  public state: IAppNewState = {
    error: undefined,
  };

  public componentDidMount() {
    const {
      chartID,
      fetchChartVersions,
      getBindings,
      getChartVersion,
      getChartValues,
      chartVersion,
    } = this.props;
    fetchChartVersions(chartID);
    getBindings();
    getChartVersion(chartID, chartVersion);
    getChartValues(chartID, chartVersion);
  }

  public componentWillReceiveProps(nextProps: IAppNewProps) {
    const { selectChartVersionAndGetFiles, chartVersion } = this.props;
    const { versions } = this.props.selected;

    if (nextProps.chartVersion !== chartVersion) {
      const cv = versions.find(v => v.attributes.version === nextProps.chartVersion);
      if (cv) {
        selectChartVersionAndGetFiles(cv);
      } else {
        throw new Error("could not find chart");
      }
    }
  }

  public render() {
    const { chartID, chartVersion, selected, bindings, deployChart, push } = this.props;
    const { version, versions } = selected;
    if (!version || !versions.length) {
      return <div>Loading</div>;
    }
    return (
      <div>
        {this.state.error && (
          <div className="container padding-v-bigger bg-action">{this.state.error}</div>
        )}
        <DeploymentForm
          chartID={chartID}
          chartVersion={chartVersion}
          selected={selected}
          bindings={bindings}
          deployChart={deployChart}
          push={push}
        />
      </div>
    );
  }
}

export default AppNew;
