/* tslint:disable:no-unused-variable */
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { HttpModule } from '@angular/http';
import { Angulartics2Module, Angulartics2GoogleAnalytics } from 'angulartics2';
import { ClipboardModule } from 'ngx-clipboard';
import { Observable } from 'rxjs/Rx';

// Shared
import { TruncatePipe } from '../shared/pipes/truncate.pipe';
import { ChartsService } from '../shared/services/charts.service';
import { ConfigService } from '../shared/services/config.service';
import { SeoService } from '../shared/services/seo.service';
import { MenuService } from '../shared/services/menu.service';

// Components
import { ChartDetailsComponent } from './chart-details.component';
import { ChartItemComponent } from '../chart-item/chart-item.component';
import { PanelComponent } from '../panel/panel.component';
import { HeaderBarComponent } from '../header-bar/header-bar.component';
import { LoaderComponent } from '../loader/loader.component';
import { ListItemComponent } from '../list-item/list-item.component';
import { ChartDetailsVersionsComponent } from './chart-details-versions/chart-details-versions.component';
import { ChartDetailsInfoComponent } from './chart-details-info/chart-details-info.component';
import { ChartDetailsReadmeComponent } from './chart-details-readme/chart-details-readme.component';
import { ChartDetailsUsageComponent } from './chart-details-usage/chart-details-usage.component';

import 'hammerjs';

// Stub
const mockData = {
  data: {
    attributes: {
      description: 'Testing the chart',
      home: 'helm.sh',
      keywords: ['artifactory'],
      maintainers: [
        {
          email: 'test@example.com',
          name: 'Test'
        }
      ],
      name: 'test',
      repo: 'incubator',
      sources: ['https://github.com/']
    },
    id: 'incubator/test',
    relationships: {
      latestChartVersion: {
        data: {
          created: '2017-02-13T04:33:57.218083521Z',
          digest:
            'eba0c51d4bc5b88d84f83d8b2ba0c5e5a3aad8bc19875598198bdbb0b675f683',
          icons: [
            {
              name: '160x160-fit',
              path: '/assets/incubator/test/4.16.0/logo-160x160-fit.png'
            }
          ],
          readme: '/assets/incubator/test/4.16.0/README.md',
          urls: [
            'https://kubernetes-charts-incubator.storage.googleapis.com/test-4.16.0.tgz'
          ],
          version: '4.16.0'
        },
        links: {
          self: '/v1/charts/incubator/test/versions/4.16.0'
        }
      }
    },
    type: 'chart'
  }
};

describe('ChartDetailsComponent', () => {
  let component: ChartDetailsComponent;
  let fixture: ComponentFixture<ChartDetailsComponent>;

  beforeEach(
    async(() => {
      TestBed.configureTestingModule({
        imports: [
          ClipboardModule,
          BrowserModule,
          Angulartics2Module,
          RouterTestingModule,
          HttpModule
        ],
        declarations: [
          ChartDetailsComponent,
          ChartDetailsVersionsComponent,
          ChartDetailsInfoComponent,
          ChartDetailsReadmeComponent,
          ChartDetailsUsageComponent,
          LoaderComponent,
          PanelComponent,
          HeaderBarComponent,
          ChartItemComponent,
          TruncatePipe,
          ListItemComponent
        ],
        providers: [
          { provide: ChartsService },
          { provide: SeoService },
          { provide: MenuService }
        ]
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(ChartDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
