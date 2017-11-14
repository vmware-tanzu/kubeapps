import { Component, OnInit } from '@angular/core';
import { ChartsService } from '../shared/services/charts.service';
import { Chart } from '../shared/models/chart';
import { SeoService } from '../shared/services/seo.service';

@Component({
  selector: 'app-chart-index',
  templateUrl: './chart-index.component.html',
  styleUrls: ['./chart-index.component.scss']
})
export class ChartIndexComponent implements OnInit {
	charts: Chart[]
  loading: boolean = true;
  totalChartsNumber: number

  constructor(
    private chartsService: ChartsService,
    private seo: SeoService
  ) {}

  ngOnInit() {
		this.loadCharts();
    this.seo.setMetaTags('index');
  }

  loadCharts(): void {
		this.chartsService.getCharts().subscribe(charts => {
      this.loading = false;
      this.charts = charts;
      this.totalChartsNumber = charts.length;
    });
  }
}
