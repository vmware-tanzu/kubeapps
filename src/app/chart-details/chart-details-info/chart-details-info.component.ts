import { Component, OnInit, Input } from '@angular/core';
import { ChartsService } from '../../shared/services/charts.service';
import { Chart } from '../../shared/models/chart';
import { Maintainer } from '../../shared/models/maintainer';
import { ChartVersion } from '../../shared/models/chart-version';
import * as urljoin from 'url-join';

@Component({
  selector: 'app-chart-details-info',
  templateUrl: './chart-details-info.component.html',
  styleUrls: ['./chart-details-info.component.scss']
})
export class ChartDetailsInfoComponent implements OnInit {
  @Input() chart: Chart;
  @Input() currentVersion: ChartVersion;
  versions: ChartVersion[];
  constructor(private chartsService: ChartsService) {}

  ngOnInit() {
    this.loadVersions(this.chart);
  }

  get sources() {
    return this.chart.attributes.sources || [];
  }

  get sourceUrl(): string {
    var chartSource = this.chart.attributes.repo.source;
    if (!chartSource) return;

    return urljoin(chartSource, this.chart.attributes.name);
  }

  get maintainers(): Maintainer[] {
    return this.chart.attributes.maintainers || [];
  }

  get sourceName(): string {
    var parser = document.createElement('a');
    parser.href = this.chart.attributes.repo.source;
    return parser.hostname;
  }

  loadVersions(chart: Chart): void {
    this.chartsService
      .getVersions(chart.attributes.repo.name, chart.attributes.name)
      .subscribe(versions => {
        this.versions = versions;
      });
  }

  maintainerUrl(maintainer: Maintainer): string {
    if (this.chart.attributes.repo.source.match(/github.com/)) {
      return `https://github.com/${maintainer.name}`;
    } else {
      return `mailto:${maintainer.email}`;
    }
  }
}
