import { Router, ActivatedRoute, Params } from '@angular/router';
import { Component, OnInit, ElementRef, ViewChild } from '@angular/core';
import { DeploymentsService } from '../shared/services/deployments.service';
import { Deployment } from '../shared/models/deployment';
import { Chart } from '../shared/models/chart';
import { SeoService } from '../shared/services/seo.service';
import { ConfigService } from '../shared/services/config.service';
import { DomSanitizer } from '@angular/platform-browser';
import { MatIconRegistry } from '@angular/material';

@Component({
  selector: 'app-deployment',
  templateUrl: './deployment.component.html',
  styleUrls: ['./deployment.component.scss'],
  viewProviders: [MatIconRegistry]
})
export class DeploymentComponent implements OnInit {
  deployment: Deployment;
  resources = [];
  loading: boolean = true;
  name: String = '';

  constructor(
    private deploymentsService: DeploymentsService,
    private router: Router,
    private route: ActivatedRoute,
    private seo: SeoService,
    private config: ConfigService,
    private mdIconRegistry: MatIconRegistry,
    private sanitizer: DomSanitizer
  ) {}

  ngOnInit() {
    const icons = [
      'layers',
      'schedule',
      'web-asset',
      'info-outline',
      'arrow-back'
    ];

    icons.forEach(icon => {
      this.mdIconRegistry.addSvgIcon(
        icon,
        this.sanitizer.bypassSecurityTrustResourceUrl(
          `/assets/icons/${icon}.svg`
        )
      );
    });
    // Do not show the page if the feature is not enabled
    if (!this.config.releasesEnabled) {
      return this.router.navigate(['/404']);
    }

    this.route.params.forEach((params: Params) => {
      this.name = params['deploymentName'];
      this.loadDeployment(params['deploymentName']);
    });
  }

  loadDeployment(deploymentName: string): void {
    this.deploymentsService
      .getDeployment(deploymentName)
      .finally(() => {
        this.loading = false;
      })
      .subscribe(deployment => {
        this.deployment = deployment;
        this.resources = this.deploymentsService.loadResources(deployment);
        this.seo.setMetaTags('deploymentDetails', {
          name: this.name,
          description: this.deployment.attributes.chartName
        });
      });
  }

  deploymentDeleted(event) {
    if (event.state == 'deleted') {
      return this.router.navigate(['/deployments']);
    }
  }

  getIconUrl(): string {
    return this.deployment.attributes.chartIcon
      ? this.deployment.attributes.chartIcon
      : '/assets/images/placeholder.png';
  }
}
