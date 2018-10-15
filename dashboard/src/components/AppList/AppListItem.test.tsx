import { shallow } from "enzyme";
import * as React from "react";
import { Link } from "react-router-dom";

import { IAppOverview } from "../../shared/types";
import { PreformattedCard } from "../Card";
import AppListItem from "./AppListItem";

it("renders a app item", () => {
  const wrapper = shallow(
    <AppListItem
      app={
        {
          namespace: "default",
          releaseName: "foo",
          status: "DEPLOYED",
          version: "1.0.0",
        } as IAppOverview
      }
    />,
  );
  const card = wrapper.find(PreformattedCard).shallow();
  expect(card.find(Link).props().title).toBe("foo");
  expect(card.find(Link).props().to).toBe("/apps/ns/default/foo");
  expect(card.find(".type-color-light-blue").text()).toBe("1.0.0");
  expect(card.find(".DEPLOYED").exists()).toBe(true);
  expect(card.find(".ListItem__content__info_tag-1").text()).toBe("default");
  expect(card.find(".ListItem__content__info_tag-2").text()).toBe("deployed");
});
