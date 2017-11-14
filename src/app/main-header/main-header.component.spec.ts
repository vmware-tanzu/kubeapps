/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { MainHeaderComponent } from './main-header.component';
import { Router } from '@angular/router';
import { ConfigService } from '../shared/services/config.service';
import { MenuService } from '../shared/services/menu.service';
import { HeaderBarComponent } from '../header-bar/header-bar.component';

describe('Component: MainHeader', () => {
  beforeEach(
    async(() => {
      TestBed.configureTestingModule({
        declarations: [MainHeaderComponent, HeaderBarComponent],
        imports: [],
        providers: [
          { provide: Router },
          { provide: MenuService }
        ]
      }).compileComponents();
    })
  );

  it('should create an instance', () => {
    let component = TestBed.createComponent(MainHeaderComponent);
    expect(component).toBeTruthy();
  });
});
