/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ChartDetailsInfoComponent } from './chart-details-info.component';
import { ChartsService } from '../../shared/services/charts.service';

describe('Component: ChartDetailsInfo', () => {
  beforeEach(
    async(() => {
      TestBed.configureTestingModule({
        declarations: [ChartDetailsInfoComponent],
        imports: [],
        providers: [{ provide: ChartsService }],
        schemas: [NO_ERRORS_SCHEMA]
      }).compileComponents();
    })
  );
  it('should create an instance', () => {
    let component = TestBed.createComponent(ChartDetailsInfoComponent);
    expect(component).toBeTruthy();
  });
});
