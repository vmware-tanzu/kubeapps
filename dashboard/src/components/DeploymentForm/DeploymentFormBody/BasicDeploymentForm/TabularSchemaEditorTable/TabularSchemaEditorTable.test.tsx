// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import TabularSchemaEditorTable, {
  TabularSchemaEditorTableProps,
} from "./TabularSchemaEditorTable";

jest.useFakeTimers();

const defaultProps = {
  columns: [],
  data: [],
  globalFilter: "",
  isLoading: false,
  saveAllChanges: jest.fn(),
  setGlobalFilter: jest.fn(),
} as TabularSchemaEditorTableProps;

it("should render all the components", () => {
  const wrapper = mount(<TabularSchemaEditorTable {...defaultProps} />);
  expect(wrapper.find(".table-control")).toExist();
  expect(wrapper.find(".pagination-buttons")).toHaveLength(2);
  expect(wrapper.find("thead")).toExist();
  expect(wrapper.find("tbody")).toExist();
  expect(wrapper.find("tfoot")).toExist();
});
