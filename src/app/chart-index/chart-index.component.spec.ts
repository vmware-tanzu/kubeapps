/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { Router } from '@angular/router';
import { ChartIndexComponent } from './chart-index.component';
import { ChartListComponent } from '../chart-list/chart-list.component';
import { ChartItemComponent } from '../chart-item/chart-item.component';
import { LoaderComponent } from '../loader/loader.component';
import { PanelComponent } from '../panel/panel.component';
import { HeaderBarComponent } from '../header-bar/header-bar.component';
import { MainHeaderComponent } from '../main-header/main-header.component';
import { ChartsService } from '../shared/services/charts.service';
import { ConfigService } from '../shared/services/config.service';
import { SeoService } from '../shared/services/seo.service';
import { MenuService } from '../shared/services/menu.service';

describe('Component: ChartIndex', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [],
      declarations: [
        ChartIndexComponent,
        ChartListComponent,
        ChartItemComponent,
        LoaderComponent,
        PanelComponent,
        HeaderBarComponent,
        MainHeaderComponent
      ],
      providers: [
        ConfigService,
        MenuService,
        { provide: ChartsService },
        { provide: SeoService },
        { provide: Router }
      ],
      schemas: [NO_ERRORS_SCHEMA]
    }).compileComponents();
  });

  it('should create an instance', () => {
    let component = TestBed.createComponent(ChartIndexComponent);
    expect(component).toBeTruthy();
  });
});
