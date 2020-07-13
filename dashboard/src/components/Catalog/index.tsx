import * as React from "react";

import { ICatalogProps } from "./Catalog";

const Catalog = React.lazy(() => import("./Catalog"));
const CatalogV2 = React.lazy(() => import("./Catalog.v2"));

interface ICatalogSelectorProps extends ICatalogProps {
  UI: string;
}

const CatalogSelector: React.FC<ICatalogSelectorProps> = props => (
  <React.Suspense fallback={null}>
    {props.UI === "clarity" ? <CatalogV2 {...props} /> : <Catalog {...props} />}
  </React.Suspense>
);

export default CatalogSelector;
