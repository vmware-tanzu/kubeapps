// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import Column from ".";

const randomColumn = () => Math.floor(Math.random() * Math.floor(11)) + 1;

const randomColumns = (numberOfColumns: number) =>
  [...Array(numberOfColumns).keys()].map(() => randomColumn());

describe(Column, () => {
  describe("Normal rendering", () => {
    it("renders the column with auto when any property is passed", () => {
      const wrapper = mountWrapper(defaultStore, <Column>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName("clr-col");
    });

    it("adds the correct role when it's a list item", () => {
      const wrapper = mountWrapper(defaultStore, <Column isListItem>Test</Column>);

      expect(wrapper.find(Column).childAt(0).prop("role")).toBe("listitem");
    });
  });

  describe("Fixed properties", () => {
    it("adds the clr-col-X class when span is a number", () => {
      const span = 6;
      const wrapper = mountWrapper(defaultStore, <Column span={span}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-${span}`);
    });

    it("adds the clr-offset-X class offset when offset is a number", () => {
      const offset = 6;
      const wrapper = mountWrapper(defaultStore, <Column offset={offset}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-${offset}`);
    });

    it("adds clr-col and clr-offset", () => {
      const span = 6;
      const offset = 6;
      const wrapper = mountWrapper(
        defaultStore,
        <Column span={span} offset={offset}>
          Test
        </Column>,
      );

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-${span}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-${offset}`);
    });
  });

  describe("Responsive Properties", () => {
    it("adds the clr-col responsive columns when span is an array of 1 element", () => {
      const span = randomColumns(1);
      const wrapper = mountWrapper(defaultStore, <Column span={span}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-sm-${span[0]}`);
    });

    it("adds the clr-col responsive columns when span is an array of 2 element", () => {
      const span = randomColumns(2);
      const wrapper = mountWrapper(defaultStore, <Column span={span}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-md-${span[1]}`);
    });

    it("adds the clr-col responsive columns when span is an array of 3 element", () => {
      const span = randomColumns(3);
      const wrapper = mountWrapper(defaultStore, <Column span={span}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-md-${span[1]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-lg-${span[2]}`);
    });

    it("adds the clr-col responsive columns when span is an array of 4 element", () => {
      const span = randomColumns(4);
      const wrapper = mountWrapper(defaultStore, <Column span={span}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-md-${span[1]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-lg-${span[2]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-xl-${span[3]}`);
    });

    it("adds the clr-col responsive columns and ignore undefined responsive breakpoints", () => {
      const span = randomColumns(5);
      const wrapper = mountWrapper(defaultStore, <Column span={span}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-md-${span[1]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-lg-${span[2]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-xl-${span[3]}`);
    });

    it("adds the clr-offset responsive offset when offset is an array of 1 element", () => {
      const offset = randomColumns(1);
      const wrapper = mountWrapper(defaultStore, <Column offset={offset}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-sm-${offset[0]}`);
    });

    it("adds the clr-offset responsive offset when offset is an array of 2 element", () => {
      const offset = randomColumns(2);
      const wrapper = mountWrapper(defaultStore, <Column offset={offset}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-md-${offset[1]}`);
    });

    it("adds the clr-offset responsive offset when offset is an array of 3 element", () => {
      const offset = randomColumns(3);
      const wrapper = mountWrapper(defaultStore, <Column offset={offset}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-md-${offset[1]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-lg-${offset[2]}`);
    });

    it("adds the clr-offset responsive offset when offset is an array of 4 element", () => {
      const offset = randomColumns(4);
      const wrapper = mountWrapper(defaultStore, <Column offset={offset}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-md-${offset[1]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-lg-${offset[2]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-xl-${offset[3]}`);
    });

    it("adds the clr-offset responsive offset and ignore undefined responsive breakpoints", () => {
      const offset = randomColumns(5);
      const wrapper = mountWrapper(defaultStore, <Column offset={offset}>Test</Column>);

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-md-${offset[1]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-lg-${offset[2]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-xl-${offset[3]}`);
    });

    it("combine both responsive properties", () => {
      const offset = randomColumns(5);
      const span = randomColumns(5);
      const wrapper = mountWrapper(
        defaultStore,
        <Column span={span} offset={offset}>
          Test
        </Column>,
      );

      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-sm-${span[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-md-${span[1]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-lg-${span[2]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-col-xl-${span[3]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-sm-${offset[0]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-md-${offset[1]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-lg-${offset[2]}`);
      expect(wrapper.find(Column).childAt(0)).toHaveClassName(`clr-offset-xl-${offset[3]}`);
    });
  });
});
