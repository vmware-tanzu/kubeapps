import { Component, OnInit, Input } from '@angular/core';
import { ChartVersion } from '../../shared/models/chart-version';
import { ChartAttributes } from '../../shared/models/chart';

@Component({
  selector: 'app-chart-details-versions',
  templateUrl: './chart-details-versions.component.html',
  styleUrls: ['./chart-details-versions.component.scss']
})
export class ChartDetailsVersionsComponent implements OnInit {
  @Input() versions: ChartVersion[]
  @Input() currentVersion: ChartVersion
  showAllVersions: boolean
  constructor() { }

  ngOnInit() { }

  goToVersionUrl(version: ChartVersion): string {
    let chart: ChartAttributes = version.relationships.chart.data
    return `/charts/${chart.repo.name}/${chart.name}/${version.attributes.version}`;
  }

  isSelected(version: ChartVersion): boolean {
    return this.currentVersion && version.attributes.version == this.currentVersion.attributes.version;
  }

  showMoreLink(): boolean {
    return this.versions && this.versions.length > 5 && !this.showAllVersions;
  }

  setShowAllVersions() {
    this.showAllVersions = true;
  }

  shownVersions(versions: ChartVersion[]): ChartVersion[] {
    if (this.versions) {
      return this.showAllVersions ? this.versions : this.versions.slice(0, 5);
    }
    return [];
  }
}
