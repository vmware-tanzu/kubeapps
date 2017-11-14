/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ChartItemComponent } from './chart-item.component';
import { ConfigService } from '../shared/services/config.service';

describe('Component: ChartItem', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [],
      declarations: [ChartItemComponent],
      providers: [ConfigService],
      schemas: [NO_ERRORS_SCHEMA]
    }).compileComponents();
  });

  it('should create an instance', () => {
    let component = TestBed.createComponent(ChartItemComponent);
    expect(component).toBeTruthy();
  });
});
