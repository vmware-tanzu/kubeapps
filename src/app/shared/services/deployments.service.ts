import { Injectable } from '@angular/core';
import { Deployment } from '../models/deployment';
import { ConfigService } from './config.service';

import { Observable } from 'rxjs';
import 'rxjs/add/operator/switchMap';
import 'rxjs/add/operator/find';
import 'rxjs/add/operator/map';

import { Http, Response } from '@angular/http';

/* TODO, This is a mocked class. */
@Injectable()
export class DeploymentsService {
  hostname: string;

  constructor(
    private http: Http,
    private config: ConfigService
  ) {
    this.hostname = config.backendHostname;
  }

  getDeployments(): Observable<Deployment[]> {
      return this.http.get(`${this.hostname}/v1/releases`, {withCredentials: true})
                    .map((response) => {
                      return this.extractData(response, [])
                    }).catch(this.handleError);
  }

  getDeployment(deploymentName: string): Observable<Deployment> {
      return this.http.get(`${this.hostname}/v1/releases/${deploymentName}`, {withCredentials: true})
                    .map((response) => {
                      return this.extractData(response, [])
                    }).catch(this.handleError);
  }

  installDeployment(chartID: string, version: string, namespace: string): Observable<Deployment> {
      var params = { "chartId": chartID, "chartVersion": version, "namespace": namespace }
      return this.http.post(`${this.hostname}/v1/releases`, params, {withCredentials: true})
                    .map(this.extractData)
                    .catch(this.handleError);
  }

  deleteDeployment(deploymentName: string): Observable<Deployment> {
    return this.http.delete(`${this.hostname}/v1/releases/${deploymentName}`, {withCredentials: true})
                    .map(this.extractData)
                    .catch(this.handleError);
  }

  private extractData(res: Response, fallback = {}) {
    let body = res.json();
    var data = body.data;
    if (!data) {
        return fallback;
    }
    var attributes = data.attributes;
    if (attributes) {
      attributes.urls = [];
      var resources = attributes.resources;
      if (resources) {
        var parsedResources = this.loadResources(data);
        parsedResources.forEach(x => {
          if (x.name == 'Service') {
              x.services.forEach(svc => {
                attributes.urls = attributes.urls.concat(this.svcToURLs(svc));
              })
          }
        })
      }
    }
    return data || fallback;
  }

  /**
   * Take a service status and try to assemble urls out of it
   * @param portSpec parsed SVC line
   * @param ret out parameter to store urls in
   */
  private svcToURLs(portSpec) {
    // pattern is EXT_PORT:INT_PORT/TCP
    const portsRE = /^(\d+):(\d+)\/TCP$/;
    // match ip4/6 ips, as opposed to <pending>
    const ipRE = /^[\d\.:]+$/;
    const EXT_IP = "EXTERNAL-IP";

    var ret = [];
    var ports = portSpec['PORT(S)'].split(",");
    ports.forEach(port => {
      var extIP = portSpec[EXT_IP];
      // only look at services with valid external IP
      if (!ipRE.exec(extIP))
        return;
      var portMatch = portsRE.exec(port);
      if (portMatch) {
        var protocol = portMatch[1] == '443' ? 'https' : 'http';
        ret.push(`${protocol}://${extIP}:${portMatch[1]}`);
      }
    })
    return ret;
  }

  /**
   * Prepare the resources for displaying in the UI.
   *
   * TODO: In the future, the backend will provide this information
   */
  loadResources(deployment: Deployment): any {
    let elements = deployment.attributes.resources.split('=='),
      resources = [];

    // Remove first element
    elements.shift();

    // Regex
    let nameRegex = /^> [\w\d\s\/]+\/(\w+)+/;

    elements.forEach(el => {
      let lines = el.split("\n");

      // Name
      let name = nameRegex.exec(lines.shift())[1];
      let headers = lines.shift().split(/\s+/);
      let services = [];

      // Remaining lines
      lines.forEach(line => {
        if (line !== '') {
          let values = line.split(/\s+/);
          let service = {};

          values.forEach((value, i) => {
            service[headers[i]] = value;
          });

          // Add to the array
          services.push(service);
        }
      });

      // Build the resource
      resources.push({ name, services });
    });

    return resources;
  }

  private handleError (error: any) {
    error = error.json();
    let errMsg = (error.message) ? error.message :
      error.status ? `${error.status} - ${error.statusText}` : 'Server error';
    console.error(errMsg); // log to console instead
    return Observable.throw(errMsg);
  }
}
