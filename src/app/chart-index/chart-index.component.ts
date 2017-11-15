import { Component, OnInit, OnDestroy } from '@angular/core';
import { ChartsService } from '../shared/services/charts.service';
import { Chart } from '../shared/models/chart';
import { SeoService } from '../shared/services/seo.service';
import { IntervalObservable } from 'rxjs/observable/IntervalObservable';

@Component({
  selector: 'app-chart-index',
  templateUrl: './chart-index.component.html',
  styleUrls: ['./chart-index.component.scss']
})
export class ChartIndexComponent implements OnInit, OnDestroy {
	charts: Chart[]
  loading: boolean = true;
  apiNotReady: boolean = false;
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
		this.chartsService.getCharts().subscribe(c => this.setCharts(c), () => {
      this.loading = false;
      this.apiNotReady = true;
      IntervalObservable.create(2000)
        .takeWhile(() => this.apiNotReady === true)
        .subscribe(() => {
          this.chartsService.getCharts().subscribe(c => this.setCharts(c));
        })
    });
  }

  setCharts(charts: Chart[]) {
    this.loading = false;
    this.apiNotReady = false;
    this.charts = charts;
    this.totalChartsNumber = charts.length;
  }

  ngOnDestroy() {
    // This ensures the IntervalObservable is cancelled
    this.apiNotReady = false;
  }
}
