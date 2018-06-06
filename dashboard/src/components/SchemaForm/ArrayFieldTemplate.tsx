import * as React from "react";
import { ArrayFieldTemplateProps } from "react-jsonschema-form";

// adapted from https://github.com/mozilla-services/react-jsonschema-form/blob/master/playground/samples/customArray.js
const ArrayFieldTemplate: React.SFC<ArrayFieldTemplateProps> = props => {
  const { className, title, items, canAdd, onAddClick } = props;
  return (
    <div className={`${className} margin-b-normal`}>
      <label>{title}</label>
      {items &&
        items.map(element => (
          <div key={element.index}>
            <div className="margin-b-normal">{element.children}</div>
            {element.hasMoveDown && (
              <button
                type="button"
                className="button button-small"
                onClick={element.onReorderClick(element.index, element.index + 1)}
              >
                Down
              </button>
            )}
            {element.hasMoveUp && (
              <button
                type="button"
                className="button button-small"
                onClick={element.onReorderClick(element.index, element.index - 1)}
              >
                Up
              </button>
            )}
            <button
              type="button"
              className="button button-danger button-small"
              onClick={element.onDropIndexClick(element.index)}
            >
              Delete
            </button>
            <hr />
          </div>
        ))}

      {canAdd && (
        <div className="row">
          <button className="button button-primary button-small" type="button" onClick={onAddClick}>
            Add Item
          </button>
        </div>
      )}
    </div>
  );
};

export default ArrayFieldTemplate;
