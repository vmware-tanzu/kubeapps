import context from "jest-plugin-context";

import ConfirmDialog from ".";
import itBehavesLike from "../../shared/specs";

context("when loading is true", () => {
  const props = {
    loading: true,
  };

  itBehavesLike("aLoadingComponent", { component: ConfirmDialog, props });
});
