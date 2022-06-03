// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// import { CdsButton } from "@cds/react/button";
// import { CdsIcon } from "@cds/react/icon";
// import actions from "actions";
// import { useState } from "react";
// import { useDispatch, useSelector } from "react-redux";
// import { IStoreState } from "shared/types";

// TODO(agamez): the refresh functionallity is currently not implemented/supported in the new Repositories API. Decide whether removing it or not
export function AppRepoRefreshAllButton() {
  // const [refreshing, setRefreshing] = useState(false);
  // const { repos } = useSelector((state: IStoreState) => state.repos);
  // const dispatch = useDispatch();
  // const handleResyncAllClick = async () => {
  //   // Fake timeout to show progress
  //   // TODO(andresmgot): Ideally, we should show the progress of the sync but we don't
  //   // have that info yet: https://github.com/vmware-tanzu/kubeapps/issues/153
  //   setRefreshing(true);
  //   setTimeout(() => setRefreshing(false), 500);
  //   if (repos) {
  //     const repoObjects = repos.map(repo => repo.packageRepoRef);
  //     dispatch(actions.repos.resyncAllRepos(repoObjects));
  //   }
  // };
  // return (
  //   <div className="refresh-all-button">
  //     <CdsButton disabled={refreshing} action="outline" onClick={handleResyncAllClick}>
  //       <CdsIcon shape="refresh" /> {refreshing ? "Refreshing" : "Refresh All"}
  //     </CdsButton>
  //   </div>
  // );
}
