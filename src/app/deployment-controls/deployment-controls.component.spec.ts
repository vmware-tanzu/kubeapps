/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { DeploymentControlsComponent } from './deployment-controls.component';
import { DeploymentsService } from '../shared/services/deployments.service';
import { DialogsService } from '../shared/services/dialogs.service';
import { ConfigService } from '../shared/services/config.service';

describe('Component: DeploymentControls', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [],
      declarations: [DeploymentControlsComponent],
      providers: [
        { provide: DeploymentsService },
        { provide: DialogsService },
        ConfigService
      ]
    }).compileComponents();
  });

  it('should create an instance', () => {
    let component = TestBed.createComponent(DeploymentControlsComponent);
    expect(component).toBeTruthy();
  });
});
