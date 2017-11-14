/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { Router } from '@angular/router';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ChartDetailsUsageComponent } from './chart-details-usage.component';
import { ConfigService } from '../../shared/services/config.service';
import { DeploymentsService } from '../../shared/services/deployments.service';
import { DialogsService } from '../../shared/services/dialogs.service';

describe('Component: ChartDetailsUsage', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [],
      declarations: [ChartDetailsUsageComponent],
      providers: [
        DialogsService,
        ConfigService,
        { provide: DeploymentsService },
        { provide: Router }
      ],
      schemas: [NO_ERRORS_SCHEMA]
    }).compileComponents();
  });

  it('should create an instance', () => {
    let component = TestBed.createComponent(ChartDetailsUsageComponent);
    expect(component).toBeTruthy();
  });
});
