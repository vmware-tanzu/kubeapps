/* tslint:disable:no-unused-variable */

import { TestBed, async } from '@angular/core/testing';
import { HeaderBarComponent } from './header-bar.component';
import { Router } from '@angular/router';
import { ConfigService } from '../shared/services/config.service';
import { MenuService } from '../shared/services/menu.service';

describe('Component: HeaderBar', () => {
  beforeEach(
    async(() => {
      TestBed.configureTestingModule({
        declarations: [HeaderBarComponent],
        imports: [],
        providers: [
          { provide: Router },
          { provide: MenuService }
        ]
      }).compileComponents();
    })
  );

  it('should create an instance', () => {
    let component = TestBed.createComponent(HeaderBarComponent);
    expect(component).toBeTruthy();
  });
});
