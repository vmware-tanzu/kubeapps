import * as React from "react";
import { IServiceBinding } from "../../shared/ServiceBinding";

export function bindingDetail(selectedBinding: IServiceBinding | undefined) {
  let bindingDetailDiv = <div />;
  if (selectedBinding) {
    const {
      instanceRef,
      secretName,
      secretDatabase,
      secretHost,
      secretPassword,
      secretPort,
      secretUsername,
    } = selectedBinding.spec;

    const statuses: Array<[string, string | undefined]> = [
      ["Instance", instanceRef.name],
      ["Secret", secretName],
      ["Database", secretDatabase],
      ["Host", secretHost],
      ["Password", secretPassword],
      ["Port", secretPort],
      ["Username", secretUsername],
    ];

    bindingDetailDiv = (
      <dl className="container margin-normal">
        {statuses.map(statusPair => {
          const [key, value] = statusPair;
          return [
            <dt key={key}>{key}</dt>,
            <dd key={value}>
              <code>{value}</code>
            </dd>,
          ];
        })}
      </dl>
    );
  }
  return bindingDetailDiv;
}

export function bindingOptions(
  bindings: IServiceBinding[],
  selectedBinding: IServiceBinding | undefined,
) {
  return (
    <div>
      <option key="none" value="none">
        {" "}
        -- Select one --
      </option>
      {bindings.map(b => (
        <option
          key={b.metadata.name}
          selected={b.metadata.name === (selectedBinding && selectedBinding.metadata.name)}
          value={b.metadata.name}
        >
          {b.metadata.name}
        </option>
      ))}
    </div>
  );
}
