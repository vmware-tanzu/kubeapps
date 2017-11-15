import { Component, OnInit, ViewEncapsulation } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { ConfigService } from '../shared/services/config.service';
import { MenuService } from '../shared/services/menu.service';
import { DomSanitizer } from '@angular/platform-browser';
import { MatIconRegistry } from '@angular/material';
import { MatSnackBar } from '@angular/material';
import { CookieService } from 'ngx-cookie';
import { AuthService } from '../shared/services/auth.service';

@Component({
  selector: 'app-header-bar',
  templateUrl: './header-bar.component.html',
  styleUrls: ['./header-bar.component.scss'],
  encapsulation: ViewEncapsulation.None,
  viewProviders: [MatIconRegistry],
  inputs: ['showSearch', 'transparent']
})
export class HeaderBarComponent implements OnInit {
  // Whether or not the Monocular server requires authentication
  public loggedIn: boolean = false;
  // public user
  public user: any = {};
  // Show search form by default
  public showSearch: boolean = true;
  // Set the background as transparent
  public transparent: boolean = false;
  // Check if  the menu is opened
  public openedMenu: boolean = false;
  // Config

  constructor(
    private router: Router,
    public config: ConfigService,
    private menuService: MenuService,
    private mdIconRegistry: MatIconRegistry,
    private sanitizer: DomSanitizer,
    private cookieService: CookieService,
    private authService: AuthService,
  ) {}

  ngOnInit() {
    // Set the icon
    this.mdIconRegistry.addSvgIcon(
      'menu',
      this.sanitizer.bypassSecurityTrustResourceUrl('/assets/icons/menu.svg')
    );
    this.mdIconRegistry.addSvgIcon(
      'close',
      this.sanitizer.bypassSecurityTrustResourceUrl('/assets/icons/close.svg')
    );
    this.mdIconRegistry.addSvgIcon(
      'search',
      this.sanitizer.bypassSecurityTrustResourceUrl('/assets/icons/search.svg')
    );
    this.mdIconRegistry.addSvgIcon(
      'github',
      this.sanitizer.bypassSecurityTrustResourceUrl('/assets/icons/github.svg')
    );

    this.authService.loggedIn().subscribe(loggedIn => { this.loggedIn = loggedIn; });

    let userClaims = this.cookieService.get("ka_claims");
    if (userClaims) {
      this.user = JSON.parse(atob(userClaims));
    }
  }

  logout() {
    this.authService.logout().subscribe(() => {
      window.location.reload();
    });
  }

  searchCharts(input: HTMLInputElement): void {
    // Empty query
    if (input.value === '') {
      this.router.navigate(['/charts']);
    } else {
      let navigationExtras: NavigationExtras = {
        queryParams: { q: input.value }
      };
      this.router.navigate(['/charts'], navigationExtras);
    }
  }

  openMenu() {
    // Open the menu
    this.openedMenu = !this.openedMenu;
    this.menuService.toggleMenu();
  }
}
