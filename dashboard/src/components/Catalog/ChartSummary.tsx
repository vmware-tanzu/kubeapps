import ChartMaintainers from "components/ChartView/ChartMaintainers";
import { IChartAttributes, IChartVersion } from "shared/types";

interface IChartSummaryProps {
  version: IChartVersion;
  chartAttrs: IChartAttributes;
}

function isKubernetesCharts(repoURL: string) {
  return (
    repoURL === "https://kubernetes-charts.storage.googleapis.com" ||
    repoURL === "https://kubernetes-charts-incubator.storage.googleapis.com"
  );
}

export default function ChartSummary({ version, chartAttrs }: IChartSummaryProps) {
  return (
    <div className="left-menu">
      {version.attributes.app_version && (
        <section className="left-menu-subsection" aria-labelledby="chartinfo-versions">
          <h5 className="left-menu-subsection-title" id="chartinfo-versions">
            App Version
          </h5>
          <div>{version.attributes.app_version}</div>
        </section>
      )}
      {chartAttrs.home && (
        <section className="left-menu-subsection" aria-labelledby="chartinfo-versions">
          <h5 className="left-menu-subsection-title" id="chartinfo-versions">
            Home
          </h5>
          <div>
            <a href={chartAttrs.home} target="_blank" rel="noopener noreferrer">
              {chartAttrs.home}
            </a>
          </div>
        </section>
      )}
      {chartAttrs.maintainers?.length > 0 && (
        <section className="left-menu-subsection" aria-labelledby="chartinfo-versions">
          <h5 className="left-menu-subsection-title" id="chartinfo-versions">
            Maintainers
          </h5>
          <div>
            <ChartMaintainers
              maintainers={chartAttrs.maintainers}
              githubIDAsNames={isKubernetesCharts(chartAttrs.repo.url)}
            />
          </div>
        </section>
      )}
      {chartAttrs.sources?.length > 0 && (
        <section className="left-menu-subsection" aria-labelledby="chartinfo-versions">
          <h5 className="left-menu-subsection-title" id="chartinfo-versions">
            Related
          </h5>
          <div>
            <ul>
              {chartAttrs.sources.map((s, i) => (
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
