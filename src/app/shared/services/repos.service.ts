import { Injectable } from '@angular/core';
import { Repo, RepoAttributes } from '../models/repo';
import { ConfigService } from './config.service';

import { Observable } from 'rxjs';
import 'rxjs/add/operator/switchMap';
import 'rxjs/add/operator/find';
import 'rxjs/add/operator/map';

import { Http, Response } from '@angular/http';

/* TODO, This is a mocked class. */
@Injectable()
export class ReposService {
  hostname: string;

  constructor(
    private http: Http,
    private config: ConfigService
  ) {
    this.hostname = config.backendHostname;
  }

  /**
   * Get all repos from the API
   *
   * @return {Observable} An observable that will an array with all repos
   */
  getRepos(): Observable<Repo[]> {
    return this.http.get(`${this.hostname}/v1/repos`)
                  .map(this.extractData)
                  .catch(this.handleError);
  }
  
  /**
   * Create a repo
   *
   * @return {Observable} An observable repo
   */
  createRepo(params: RepoAttributes): Observable<Repo> {
    return this.http.post(`${this.hostname}/v1/repos`, params, {withCredentials: true})
                  .map(this.extractData)
                  .catch(this.handleError);
  }

  /**
   * Delete a repo
   *
   * @return {Observable} An observable of the deleted repo
   */
  deleteRepo(repoName: string): Observable<Repo> {
    return this.http.delete(`${this.hostname}/v1/repos/${repoName}`, {withCredentials: true})
                    .map(this.extractData)
                    .catch(this.handleError);
  }

  private extractData(res: Response) {
    let body = res.json();
    return body.data || { };
  }

  private handleError (error: any) {
    let errMsg = (error.json().message) ? error.json().message :
      error.status ? `${error.status} - ${error.statusText}` : 'Server error';
    console.error(errMsg); // log to console instead
    return Observable.throw(errMsg);
  }

}
