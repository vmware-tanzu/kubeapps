import { AxiosRequestConfig } from "axios";
import * as qs from "qs";
import * as React from "react";
import AceEditor from "react-ace";

import { axios } from "../../shared/Auth";

import "brace/mode/json";
import "brace/mode/plain_text";
import "brace/theme/xcode";

import { IFunction } from "../../shared/types";

enum Method {
  GET = "GET",
  POST = "POST",
}
const Methods = [Method.GET, Method.POST];

enum Format {
  JSON = "JSON",
  Text = "Text",
}
const Formats = {
  [Format.JSON]: "application/json",
  [Format.Text]: "application/x-www-form-urlencoded",
};
const FormatDefaults = {
  [Format.JSON]: `{
  "foo": "bar"
}`,
  [Format.Text]: "foo=bar",
};

interface IFunctionTesterProps {
  function: IFunction;
}

interface IFunctionTesterState {
  body: string;
  dirty: boolean;
  format: Format;
  method: Method;
  response?: any;
}

class FunctionTester extends React.Component<IFunctionTesterProps, IFunctionTesterState> {
  public state: IFunctionTesterState = {
    body: FormatDefaults[Format.JSON],
    dirty: false,
    format: Format.JSON,
    method: Method.GET,
  };

  public render() {
    return (
      <div className="FunctionTester">
        <h6>Test Function</h6>
        <hr />
        <div className="row">
          <div className="col-6">
            {Methods.map(m => (
              <label htmlFor={`FunctionTester__method--${m}`} className="radio" key={m}>
                <input
                  type="radio"
                  name="method"
                  value={m}
                  id={`FunctionTester__method--${m}`}
                  checked={this.state.method === m}
                  onChange={this.handleMethodChange}
                />
                <span>{m}</span>
              </label>
            ))}
          </div>
          <div className="col-6 text-r">
            {Object.keys(Formats).map(f => (
              <label htmlFor={`FunctionTester__format--${f}`} className="radio" key={f}>
                <input
                  type="radio"
                  name="format"
                  value={f}
                  id={`FunctionTester__format--${f}`}
                  checked={this.state.format === f}
                  onChange={this.handleFormatChange}
                />
                <span>{f}</span>
              </label>
            ))}
          </div>
        </div>
        <AceEditor
          mode={this.state.format === Format.JSON ? "json" : "plain_text"}
          theme="xcode"
          name="body"
          width="100%"
          height="10em"
          onChange={this.handleBodyChange}
          value={this.state.body}
        />
        <button
          className="button button-primary button-small margin-t-normal"
          onClick={this.handleTestButtonClick}
        >
          Test
        </button>
        {this.state.response && (
          <pre>
            <code>{this.state.response}</code>
          </pre>
        )}
      </div>
    );
  }

  private handleMethodChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    this.setState({ method: Method[e.currentTarget.value] });
  };
  private handleFormatChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const format = Format[e.currentTarget.value];
    const s: Partial<IFunctionTesterState> = { format };
    if (!this.state.dirty) {
      s.body = FormatDefaults[format];
    }
    this.setState(s as IFunctionTesterState);
  };
  private handleBodyChange = (value: string) => {
    this.setState({ body: value, dirty: true });
  };
  private handleTestButtonClick = async () => {
    const { function: f } = this.props;
    const { method, format, body } = this.state;
    const url = `/api/kube/api/v1/namespaces/${f.metadata.namespace}/services/http:${
      f.metadata.name
    }:http-function-port/proxy/`;
    // Parse JSON or format query params
    const reqBody = format === Format.JSON ? JSON.parse(body) : body.replace(/\n/g, "&");
    const config: AxiosRequestConfig = {
      data: reqBody,
      headers: { "Content-Type": Formats[format] },
      method: method.toLowerCase(),
      // disable axios JSON parsing, always get a raw string back
      transformResponse: r => r,
      url,
    };
    if (method === Method.GET) {
      // Params has to be an object, so for the Text format we parse it
      config.params = format === Format.Text ? qs.parse(reqBody) : reqBody;
    }
    const res = await axios.request(config);
    this.setState({
      response: res.data,
    });
  };
}
export default FunctionTester;
