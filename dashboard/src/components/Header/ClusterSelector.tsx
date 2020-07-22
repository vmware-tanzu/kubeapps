import * as React from "react";
import { useDispatch } from "react-redux";
import Select from "react-select";

import actions from "actions";
import { IClustersState } from "reducers/cluster";

export interface IClusterSelectorProps {
  clusters: IClustersState;
  onChange: (cluster: string) => void;
}

const ClusterSelector: React.FC<IClusterSelectorProps> = props => {
  const clusters = Object.keys(props.clusters.clusters);
  const options = clusters.length > 0 ? clusters.map(c => ({ value: c, label: c })) : [];
  const dispatch = useDispatch();
  const handleClusterChange = (option: any) => {
    dispatch(actions.namespace.fetchNamespaces(option.value));
    props.onChange(option.value);
  };

  return (
    <div className="NamespaceSelector margin-r-normal">
      <label className="NamespaceSelector__label type-tiny">CLUSTER</label>
      <Select
        className="NamespaceSelector__select type-small"
        value={props.clusters.currentCluster}
        options={options}
        multi={false}
        onChange={handleClusterChange}
        clearable={false}
      />
    </div>
  );
};

export default ClusterSelector;
