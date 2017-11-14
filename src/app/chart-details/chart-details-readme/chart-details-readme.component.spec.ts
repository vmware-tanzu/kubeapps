/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ChartDetailsReadmeComponent } from './chart-details-readme.component';
import { ChartsService } from '../../shared/services/charts.service';

describe('Component: ChartDetailsReadme', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [],
      declarations: [ChartDetailsReadmeComponent],
      providers: [{ provide: ChartsService }],
      schemas: [NO_ERRORS_SCHEMA]
    }).compileComponents();
  });

  it('should create an instance', () => {
    let component = TestBed.createComponent(ChartDetailsReadmeComponent);
    expect(component).toBeTruthy();
  });
});
