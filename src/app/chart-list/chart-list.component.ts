import { Component, OnInit, Input } from '@angular/core';
import { Chart } from '../shared/models/chart';

@Component({
  selector: 'app-chart-list',
  templateUrl: './chart-list.component.html',
  styleUrls: ['./chart-list.component.scss']
})
export class ChartListComponent implements OnInit {
  @Input() charts: Chart[];

  constructor() {}

  ngOnInit() {}
}
