// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { compareItems, rankItem } from "@tanstack/match-sorter-utils";
import { FilterFn, SortingFn, sortingFns } from "@tanstack/react-table";

export const fuzzyFilter: FilterFn<any> = (row, columnId, value, addMeta) => {
  // Rank the item
  const itemRank = rankItem(row.getValue(columnId), value);

  // The search in nested rows should be supported by the library, but it's not yet
  // We will mark a parent row as a match if any of its children match
  // https://github.com/vmware-tanzu/kubeapps/issues/5437
  row.subRows?.forEach((subRow: any) => {
    const subRowRank = rankItem(subRow.getValue(columnId), value);
    itemRank.passed = subRowRank.passed || itemRank.passed;
  });

  // Store the itemRank info
  addMeta({ itemRank });
  // Return if the item should be filtered in/out
  return itemRank.passed;
};

export const fuzzySort: SortingFn<any> = (rowA: any, rowB: any, columnId: any) => {
  let dir = 0;
  // Only sort by rank if the column has ranking information
  if (rowA.columnFiltersMeta[columnId]) {
    dir = compareItems(
      rowA.columnFiltersMeta[columnId]?.itemRank,
      rowB.columnFiltersMeta[columnId]?.itemRank,
    );
  }
  // Provide an alphanumeric fallback for when the item ranks are equal
  return dir === 0 ? sortingFns.alphanumeric(rowA, rowB, columnId) : dir;
};
