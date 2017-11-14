import { Angulartics2GoogleAnalytics } from 'angulartics2';
import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { MenuService } from './shared/services/menu.service';
import { ChartsService } from './shared/services/charts.service';
import { ConfigService } from './shared/services/config.service';
import { SeoService } from './shared/services/seo.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  providers: [MenuService, ChartsService]
})
export class AppComponent {
  // Show the global menu
  public showMenu: boolean = false;
  // Config
  public config;

  constructor(
    angulartics2GoogleAnalytics: Angulartics2GoogleAnalytics,
    config: ConfigService,
    private menuService: MenuService,
    private router: Router,
    private seo: SeoService
  ) {
    menuService.menuOpen$.subscribe(show => {
      this.showMenu = show;
    });

    // Hide menu when user changes the route
    router.events.subscribe(() => {
      menuService.hideMenu();
    });
    this.config = config;
  }
}
