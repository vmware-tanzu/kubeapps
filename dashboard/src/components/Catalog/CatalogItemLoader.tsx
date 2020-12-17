import React from "react";
import ContentLoader from "react-content-loader";

export interface ICatalogItemLoader {
  loaded?: boolean;
  children?: any;
  innerProps?: string;
}

function CatalogItemLoader(props: ICatalogItemLoader) {
  const { loaded, children, innerProps } = props;
  return loaded ? (
    children
  ) : (
    <ContentLoader
      speed={2}
      width={300}
      height={300}
      className="contentLoader"
      viewBox="0 0 200 200"
      backgroundColor="#dedede"
      foregroundColor="#ffffff"
      {...innerProps}
    >
      <rect x="21" y="30" rx="0" ry="0" width="200" height="10" />
      <circle cx="61" cy="78" r="20" />
      <rect x="110" y="60" rx="0" ry="0" width="98" height="5" />
      <rect x="21" y="111" rx="0" ry="0" width="200" height="19" />
      <rect x="110" y="70" rx="0" ry="0" width="98" height="5" />
      <rect x="110" y="80" rx="0" ry="0" width="98" height="5" />
      <rect x="110" y="90" rx="0" ry="0" width="98" height="5" />

      <rect x="21" y="30" rx="0" ry="0" width="200" height="10" />
      <circle cx="61" cy="78" r="20" />
      <rect x="110" y="60" rx="0" ry="0" width="98" height="5" />
      <rect x="21" y="111" rx="0" ry="0" width="200" height="19" />
      <rect x="110" y="70" rx="0" ry="0" width="98" height="5" />
      <rect x="110" y="80" rx="0" ry="0" width="98" height="5" />
      <rect x="110" y="90" rx="0" ry="0" width="98" height="5" />

      <rect x="21" y="30" rx="0" ry="0" width="200" height="10" />
      <circle cx="61" cy="78" r="20" />
      <rect x="110" y="60" rx="0" ry="0" width="98" height="5" />
      <rect x="21" y="111" rx="0" ry="0" width="200" height="19" />
      <rect x="110" y="70" rx="0" ry="0" width="98" height="5" />
      <rect x="110" y="80" rx="0" ry="0" width="98" height="5" />
      <rect x="110" y="90" rx="0" ry="0" width="98" height="5" />
    </ContentLoader>
  );
}

export default CatalogItemLoader;
