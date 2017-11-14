import { Injectable } from '@angular/core';

@Injectable()
export class ConfigService {
  // Configurable options
  // They can be overriden using assets/js/overrides.js
  backendHostname: string
  releasesEnabled: boolean
  // EO configurable options

  constructor() {
    var overrides: any = {}
    if (window["monocular"]) {
      overrides = window["monocular"]["overrides"] || {};
    }

    this.backendHostname = overrides.backendHostname || "/api";
    this.releasesEnabled = overrides.releasesEnabled;
  }
}
