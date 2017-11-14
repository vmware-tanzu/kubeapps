import { Component, OnInit, Input, ViewEncapsulation } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';

@Component({
  selector: 'app-main-header',
  templateUrl: './main-header.component.html',
  styleUrls: ['./main-header.component.scss'],
  encapsulation: ViewEncapsulation.None
})
export class MainHeaderComponent implements OnInit {
  @Input() totalChartsNumber: number
  // Store the router
  constructor(private router: Router) { }
  ngOnInit() { }

  searchCharts(input: HTMLInputElement): void {
    // Empty query
    if(input.value === ''){
      this.router.navigate(['/charts']);
    } else {
      let navigationExtras: NavigationExtras = {
        queryParams: { 'q': input.value }
      };
      this.router.navigate(['/charts'], navigationExtras);
    }
  }
}
