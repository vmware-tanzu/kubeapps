// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import AvailablePackageMaintainers from "components/PackageHeader/AvailablePackageMaintainers";
import { AvailablePackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";

interface AvailablePackageDetailExcerptProps {
  pkg?: AvailablePackageDetail;
}

function isKubernetesPackages(repoURL: string) {
  return (
    repoURL === "https://kubernetes-charts.storage.googleapis.com" ||
    repoURL === "https://kubernetes-charts-incubator.storage.googleapis.com"
  );
}

export default function AvailablePackageDetailExcerpt({ pkg }: AvailablePackageDetailExcerptProps) {
  return (
    <div className="left-menu">
      {pkg?.version?.appVersion && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-appversion"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-appversion">
            App Version
          </h5>
          <div>{pkg.version?.appVersion}</div>
        </section>
      )}
      {pkg?.version?.pkgVersion && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-pkgversion"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-pkgversion">
            Package Version
          </h5>
          <div>{pkg.version?.pkgVersion}</div>
        </section>
      )}
      {pkg?.categories && pkg?.categories?.length > 0 && pkg?.categories[0] !== "" && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-categories"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-categories">
            Categories
          </h5>
          <div>
            <ul>
              {pkg.categories.map((s, i) => (
                <li key={i}> {s} </li>
              ))}
            </ul>
          </div>
        </section>
      )}
      {pkg?.homeUrl && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-home"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-home">
            Home
          </h5>
          <div>
            <a href={pkg.homeUrl} target="_blank" rel="noopener noreferrer">
              {pkg.homeUrl}
            </a>
          </div>
        </section>
      )}
      {pkg?.maintainers && pkg?.maintainers?.length > 0 && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-maintainers"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-maintainers">
            Maintainers
          </h5>
          <div>
            <AvailablePackageMaintainers
              maintainers={pkg.maintainers}
              githubIDAsNames={isKubernetesPackages(pkg.repoUrl)}
            />
          </div>
        </section>
      )}
      {pkg?.sourceUrls && pkg?.sourceUrls?.length > 0 && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-sources"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-sources">
            Related
          </h5>
          <div>
            <ul>
              {pkg.sourceUrls.map((s, i) => (
                <li key={i}>
                  <a href={s} target="_blank" rel="noopener noreferrer">
                    {s}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </section>
      )}
    </div>
  );
}
