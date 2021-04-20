import { useEffect } from "react";
import ReactMarkdown from "react-markdown";
import { useDispatch } from "react-redux";
import remarkGfm from "remark-gfm";

import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import Alert from "components/js/Alert";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import HeadingRenderer from "./HeadingRenderer";
import LinkRenderer from "./LinkRenderer";
import TableRenderer from "./TableRenderer";

interface IChartReadmeProps {
  cluster: string;
  namespace: string;
  chartID: string;
  version: string;
  error?: string;
  readme?: string;
}

function ChartReadme({ chartID, error, cluster, namespace, readme, version }: IChartReadmeProps) {
  const dispatch = useDispatch();
  useEffect(() => {
    dispatch(actions.charts.getChartReadme(cluster, namespace, chartID, version));
  }, [dispatch, cluster, namespace, chartID, version]);

  if (error) {
    if (error.toLocaleLowerCase().includes("not found")) {
      return (
        <div className="section-not-found">
          <div>
            <CdsIcon shape="file" size="64" />
            <h4>No README found</h4>
          </div>
        </div>
      );
    }
    return <Alert theme="danger">Unable to fetch chart README: {error}</Alert>;
  }
  return (
    <LoadingWrapper
      className="margin-t-xxl"
      loadingText="Fetching application README..."
      loaded={!!readme}
    >
      {readme && (
        <div className="application-readme">
          <ReactMarkdown
            plugins={[remarkGfm]}
            components={{
              h1: HeadingRenderer,
              h2: HeadingRenderer,
              h3: HeadingRenderer,
              h4: HeadingRenderer,
              h5: HeadingRenderer,
              h6: HeadingRenderer,
              a: LinkRenderer,
              table: TableRenderer,
            }}
            skipHtml={true}
          >
            {readme}
          </ReactMarkdown>
        </div>
      )}
    </LoadingWrapper>
  );
}

export default ChartReadme;
