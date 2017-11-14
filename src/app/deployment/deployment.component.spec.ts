/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { Router, ActivatedRoute } from '@angular/router';
import { DeploymentsService } from '../shared/services/deployments.service';
import { ConfigService } from '../shared/services/config.service';
import { MenuService } from '../shared/services/menu.service';
import { DeploymentComponent } from './deployment.component';
import { LoaderComponent } from '../loader/loader.component';
import { PanelComponent } from '../panel/panel.component';
import { DeploymentControlsComponent } from '../deployment-controls/deployment-controls.component';
import { DeploymentResourceComponent } from './deployment-resource/deployment-resource.component';
import { HeaderBarComponent } from '../header-bar/header-bar.component';

describe('Component: Deployment View', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [],
      declarations: [
        DeploymentComponent,
        DeploymentControlsComponent,
        DeploymentResourceComponent,
        LoaderComponent,
        HeaderBarComponent,
        PanelComponent
      ],
      providers: [
        ConfigService,
        MenuService,
        { provide: DeploymentsService },
        { provide: ActivatedRoute },
        { provide: Router }
      ]
    }).compileComponents();
  });

  it('should create an instance', () => {
    let component = TestBed.createComponent(DeploymentComponent);
    expect(component).toBeTruthy();
  });
});
