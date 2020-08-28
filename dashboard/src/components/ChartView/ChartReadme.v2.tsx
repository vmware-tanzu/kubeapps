import React, { useEffect } from "react";
import ReactMarkdown from "react-markdown";
import { useDispatch } from "react-redux";

import actions from "actions";
import { CdsIcon } from "components/Clarity/clarity";
import Alert from "components/js/Alert";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import HeadingRenderer from "./HeadingRenderer";
import LinkRenderer from "./LinkRenderer";
import TableRenderer from "./TableRenderer";

interface IChartReadmeProps {
  namespace: string;
  chartID: string;
  version: string;
  error?: string;
  readme?: string;
}

function ChartReadme({ chartID, error, namespace, readme, version }: IChartReadmeProps) {
  const dispatch = useDispatch();
  useEffect(() => {
    dispatch(actions.charts.getChartReadme(namespace, chartID, version));
  }, [dispatch, namespace, chartID, version]);

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
    <LoadingWrapper loaded={!!readme}>
      {readme && (
        <div className="application-readme">
          <ReactMarkdown
            source={readme}
            renderers={{
              heading: HeadingRenderer,
              link: LinkRenderer,
              table: TableRenderer,
            }}
            skipHtml={true}
          />
        </div>
      )}
    </LoadingWrapper>
  );
}

export default ChartReadme;
