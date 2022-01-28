// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// CODE ADAPTED FROM: https://sghall.github.io/react-compound-slider/#/slider-demos/horizontal
import React from "react";
import { Handles, Rail, Slider as ReactSlider, Tracks } from "react-compound-slider";
import { Handle, Track } from "./components"; // example render components

const sliderStyle: React.CSSProperties = {
  margin: "1.2em",
  position: "relative",
  width: "90%",
};

const railStyle: React.CSSProperties = {
  position: "absolute",
  width: "100%",
  height: 14,
  borderRadius: 7,
  cursor: "pointer",
  backgroundColor: "rgb(155,155,155)",
};

export interface ISliderProps {
  min: number;
  max: number;
  default: number;
  step: number;
  values: number;
  sliderStyle?: React.CSSProperties;
  onChange: (values: readonly number[]) => void;
  onUpdate: (values: readonly number[]) => void;
}

class Slider extends React.Component<ISliderProps> {
  public render() {
    const { min, max, step, values, onUpdate, onChange } = this.props;
    const domain = [min, max];
    return (
      <ReactSlider
        mode={1}
        step={step}
        domain={domain}
        rootStyle={{
          ...sliderStyle,
          ...this.props.sliderStyle,
        }}
        onUpdate={onUpdate}
        onChange={onChange}
        values={[values]}
      >
        <Rail>{({ getRailProps }) => <div style={railStyle} {...getRailProps()} />}</Rail>
        <Handles>
          {({ handles, getHandleProps }) => (
            <div className="slider-handles">
              {handles.map(handle => (
                <Handle
                  key={handle.id}
                  handle={handle}
                  domain={domain}
                  getHandleProps={getHandleProps}
                />
              ))}
            </div>
          )}
        </Handles>
        <Tracks right={false}>
          {({ tracks, getTrackProps }) => (
            <div className="slider-tracks">
              {tracks.map(({ id, source, target }) => (
                <Track key={id} source={source} target={target} getTrackProps={getTrackProps} />
              ))}
            </div>
          )}
        </Tracks>
      </ReactSlider>
    );
  }
}

export default Slider;
