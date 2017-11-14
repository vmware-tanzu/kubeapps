import { Injectable } from '@angular/core';
import { ConfigService } from './config.service';
import { CookieService } from 'ngx-cookie';
import { Router } from '@angular/router';

import { Observable } from 'rxjs';
import 'rxjs/add/operator/switchMap';
import 'rxjs/add/operator/find';
import 'rxjs/add/operator/map';

import { Http, Response } from '@angular/http';

/* TODO, This is a mocked class. */
@Injectable()
export class AuthService {
  hostname: string;

  constructor(
    private http: Http,
    private config: ConfigService,
    private cookieService: CookieService,
    private router: Router,
  ) {
    this.hostname = config.backendHostname;
  }

  /**
   * Check if logged in on the API server
   * 
   * @return {Observable} An observable boolean that will be true if logged in or if auth is disabled
   */
  loggedIn(): Observable<boolean> {
    return this.http.get(`${this.hostname}/auth/verify`, {withCredentials: true})
      .map((res: Response) => { return res.ok; })
      .catch(res => {
        if (res.status == 404) {
          // If 404, authentication is disabled on the API server and we are considered logged in
          return Observable.of(true);
        } else {
          return Observable.of(false);
        }
      });
  }

  /**
   * Logs user out
   */
  logout() {
    this.cookieService.remove('ka_claims');
    this.http.delete(`${this.hostname}/auth/logout`, {withCredentials: true}).subscribe();
  }
}
