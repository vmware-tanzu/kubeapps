import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Params } from '@angular/router';
import { ChartsService } from '../shared/services/charts.service';
import { Chart } from '../shared/models/chart';
import { ChartVersion } from '../shared/models/chart-version';
import { SeoService } from '../shared/services/seo.service';
import { ConfigService } from '../shared/services/config.service';
import { DomSanitizer } from '@angular/platform-browser';

@Component({
  selector: 'app-chart-details',
  templateUrl: './chart-details.component.html',
  styleUrls: ['./chart-details.component.scss']
})
export class ChartDetailsComponent implements OnInit {
  /* This resource will be different, probably ChartVersion */
  chart: Chart;
  loading: boolean = true;
  currentVersion: ChartVersion;
  iconUrl: string;
  titleVersion: string;

  constructor(
    private route: ActivatedRoute,
    private chartsService: ChartsService,
    private config: ConfigService,
    private seo: SeoService,
    private sanitizer: DomSanitizer,
  ) {}

  ngOnInit() {
    this.route.params.forEach((params: Params) => {
      let repo = params['repo'];
      let chartName = params['chartName'];
      let chartId = `${repo}/${chartName}`;
      this.chartsService.getChart(repo, chartName).subscribe(chart => {
        this.loading = false;
        this.chart = chart;
        let version =
          params['version'] ||
          this.chart.relationships.latestChartVersion.data.version;
        this.chartsService
          .getVersion(repo, chartName, version)
          .subscribe(chartVersion => {
            this.currentVersion = chartVersion;
            this.titleVersion = this.currentVersion.attributes.app_version || '';
            this.updateMetaTags();
          });
        this.iconUrl = this.getIconUrl();
      });
    });
  }

  /**
   * Update the metatags with the name and the description of the application.
   */
  updateMetaTags(): void {
    if (this.titleVersion.length > 0) {
      this.seo.setMetaTags('chartDetailsWithVersion', {
        name: this.chart.attributes.name,
        description: this.chart.attributes.description,
        version: this.titleVersion
      });
    } else {
      this.seo.setMetaTags('chartDetails', {
        name: this.chart.attributes.name,
        description: this.chart.attributes.description
      });
    }
  }

  goToRepoUrl(): string {
    return `/charts/${this.chart.attributes.repo.name}`;
  }

  getIconUrl(): string {
    let icons = this.chart.relationships.latestChartVersion.data.icons;
    if (icons !== undefined && icons.length > 0) {
      const icon =
        this.config.backendHostname +
        icons.find(icon => icon.name === '160x160-fit').path;
      return icon;
    } else {
      return '/assets/images/placeholder.png';
    }
  }
}
