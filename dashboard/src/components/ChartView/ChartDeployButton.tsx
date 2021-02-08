import { push } from "connected-react-router";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { IChartVersion, IStoreState } from "../../shared/types";
import { app } from "../../shared/url";

export interface IChartDeployButtonProps {
  version: IChartVersion;
  namespace: string;
  kubeappsNamespace: string;
}

function ChartDeployButton(props: IChartDeployButtonProps) {
  const [clicked, setClicked] = useState(false);
  const setClickedTrue = () => setClicked(true);

  const currentCluster = useSelector((state: IStoreState) => state.clusters.currentCluster);
  const dispatch = useDispatch();

  if (clicked) {
    const newAppURL = app.apps.new(
      currentCluster,
      props.namespace,
      props.version,
      props.version.attributes.version,
      props.kubeappsNamespace,
    );
    dispatch(push(newAppURL));
  }
  return (
    <div className="ChartDeployButton text-r">
      <button className="button button-primary button-accent" onClick={setClickedTrue}>
        Deploy
      </button>
    </div>
  );
}

export default ChartDeployButton;
