/* tslint:disable:no-unused-variable */

import { NO_ERRORS_SCHEMA } from '@angular/core';
import { TestBed, async } from '@angular/core/testing';
import { DeploymentsComponent } from './deployments.component';
import { LoaderComponent } from '../loader/loader.component';
import { PanelComponent } from '../panel/panel.component';
import { HeaderBarComponent } from '../header-bar/header-bar.component';
import { Router } from '@angular/router';
import { ConfigService } from '../shared/services/config.service';
import { MenuService } from '../shared/services/menu.service';
import { DeploymentsService } from '../shared/services/deployments.service';

describe('Component: Deployments', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [],
      declarations: [
        DeploymentsComponent,
        HeaderBarComponent,
        LoaderComponent,
        PanelComponent
      ],
      providers: [
        MenuService,
        { provide: DeploymentsService },
        { provide: Router },
        { provide: ConfigService, useValue: { releasesEnabled: true } }
      ],
      schemas: [NO_ERRORS_SCHEMA]
    }).compileComponents();
  });

  it('should create an instance', () => {
    let component = TestBed.createComponent(DeploymentsComponent);
    expect(component).toBeTruthy();
  });
});
