// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import Table from ".";
import { ITableColumnProps } from "./Table";
import TableRow from "./components/TableRow";

const columns: ITableColumnProps[] = [
  {
    accessor: "uuid",
    Header: "UUID",
  },
  {
    accessor: "name",
    Header: "Name",
  },
];

const rows = [
  {
    uuid: "1khj1kjas-quhkjwa-qjkwdka-1dkjasdna",
    name: "Test",
  },
  {
    uuid: "lkashdu21jkhasudkj12n0sdahkj12kjasdj",
    name: "Second Test",
  },
];

describe(Table, () => {
  it("render the table header", () => {
    const wrapper = shallow(<Table data={rows} columns={columns} />);

    expect(wrapper.find("thead")).toExist();
    expect(wrapper.find("thead th")).toExist();

    wrapper
      .find("thead tr")
      .children()
      .forEach((th, i) => {
        expect(th).toHaveText(columns[i].Header);
      });
  });

  it("render the rows", () => {
    const wrapper = shallow(<Table data={rows} columns={columns} />);

    expect(wrapper.find("tbody")).toExist();
    expect(wrapper.find(TableRow).length).toBe(rows.length);

    wrapper
      .find("tbody")
      .children()
      .forEach((tr, rowIter) => {
        tr.children().forEach((td, colIter) => {
          const row = rows[rowIter];
          const columnAccessor = columns[colIter].accessor;
          const expectedValue = row[columnAccessor as keyof typeof row];
          expect(td).toHaveText(expectedValue);
        });
      });
  });

  it("apply the compact style based on props", () => {
    const wrapper = shallow(<Table data={rows} columns={columns} compact />);

    expect(wrapper).toHaveClassName("table-compact");
  });

  it("apply the no border style based on props", () => {
    const wrapper = shallow(<Table data={rows} columns={columns} noBorder />);

    expect(wrapper).toHaveClassName("table-noborder");
  });

  it("apply several styles based on props", () => {
    const wrapper = shallow(<Table data={rows} columns={columns} noBorder compact />);

    expect(wrapper).toHaveClassName("table-compact");
    expect(wrapper).toHaveClassName("table-noborder");
  });

  it("appends a CSS class", () => {
    const css = "myClass";
    const wrapper = shallow(<Table data={rows} columns={columns} className={css} />);

    expect(wrapper.find("table")).toHaveClassName(css);
  });
});
