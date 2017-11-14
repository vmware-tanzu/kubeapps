/* tslint:disable:no-unused-variable */
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { DeploymentResourceComponent } from './deployment-resource.component';

describe('DeploymentResourceComponent', () => {
  let component: DeploymentResourceComponent;
  let fixture: ComponentFixture<DeploymentResourceComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DeploymentResourceComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DeploymentResourceComponent);
    component = fixture.componentInstance;
    component.resource = { services: [['serviceKeyA']]}
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
