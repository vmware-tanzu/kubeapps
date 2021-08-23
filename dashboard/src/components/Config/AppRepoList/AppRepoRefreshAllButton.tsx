import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { IStoreState } from "shared/types";

export function AppRepoRefreshAllButton() {
  const [refreshing, setRefreshing] = useState(false);
  const { repos } = useSelector((state: IStoreState) => state.repos);
  const dispatch = useDispatch();

  const handleResyncAllClick = async () => {
    // Fake timeout to show progress
    // TODO(andresmgot): Ideally, we should show the progress of the sync but we don't
    // have that info yet: https://github.com/kubeapps/kubeapps/issues/153
    setRefreshing(true);
    setTimeout(() => setRefreshing(false), 500);
    if (repos) {
      const repoObjects = repos.map(repo => {
        return {
          name: repo.metadata.name,
          namespace: repo.metadata.namespace,
        };
      });
      dispatch(actions.repos.resyncAllRepos(repoObjects));
    }
  };
  return (
    <div className="refresh-all-button">
      <CdsButton action="outline" onClick={handleResyncAllClick} disabled={refreshing}>
        <CdsIcon shape="refresh" /> {refreshing ? "Refreshing" : "Refresh All"}
      </CdsButton>
    </div>
  );
}
