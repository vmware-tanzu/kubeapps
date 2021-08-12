import AvailablePackageMaintainers from "components/ChartView/AvailablePackageMaintainers";
import { AvailablePackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";

interface AvailablePackageDetailExcerptProps {
  pkg: AvailablePackageDetail;
}

/* TODO(agamez): https://github.com/kubeapps/kubeapps/issues/3165#issuecomment-882943900 */
// function isKubernetesCharts(repoURL: string) {
//   return (
//     repoURL === "https://kubernetes-charts.storage.googleapis.com" ||
//     repoURL === "https://kubernetes-charts-incubator.storage.googleapis.com"
//   );
// }

export default function AvailablePackageDetailExcerpt({ pkg }: AvailablePackageDetailExcerptProps) {
  return (
    <div className="left-menu">
      {pkg.appVersion && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-appversion"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-appversion">
            App Version
          </h5>
          <div>{pkg.appVersion}</div>
        </section>
      )}
      {pkg.pkgVersion && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-pkgversion"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-pkgversion">
            Package Version
          </h5>
          <div>{pkg.pkgVersion}</div>
        </section>
      )}
      {pkg.categories?.length > 0 && pkg.categories[0] !== "" && (
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
      {/* TODO(agamez): https://github.com/kubeapps/kubeapps/issues/3165#issuecomment-882943900 */}
      {/* {pkg.home && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-home"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-home">
            Home
          </h5>
          <div>
            <a href={pkg.home} target="_blank" rel="noopener noreferrer">
              {pkg.home}
            </a>
          </div>
        </section>
      )} */}
      {pkg.maintainers?.length > 0 && (
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
              // TODO(agamez): https://github.com/kubeapps/kubeapps/issues/3165#issuecomment-882943900
              // githubIDAsNames={isKubernetesCharts(pkg.repo.url)}
            />
          </div>
        </section>
      )}
      {/* TODO(agamez): https://github.com/kubeapps/kubeapps/issues/3165#issuecomment-882943900 */}
      {/* {pkg.sources?.length > 0 && (
        <section
          className="left-menu-subsection"
          aria-labelledby="availablePackageDetailExcerpt-sources"
        >
          <h5 className="left-menu-subsection-title" id="availablePackageDetailExcerpt-sources">
            Related
          </h5>
          <div>
            <ul>
              {pkg.sources.map((s, i) => (
                <li key={i}>
                  <a href={s} target="_blank" rel="noopener noreferrer">
                    {s}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </section>
      )} */}
    </div>
  );
}
