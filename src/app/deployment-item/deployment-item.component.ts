import { Component, OnInit, ViewChild, ElementRef } from '@angular/core';
import { Deployment } from '../shared/models/deployment';
import { ConfigService } from '../shared/services/config.service';

@Component({
  selector: 'app-deployment-item',
  templateUrl: './deployment-item.component.html',
  styleUrls: ['./deployment-item.component.scss'],
  inputs: ['deployment']
})
export class DeploymentItemComponent implements OnInit {
  public themeColor: string;
  public iconUrl: string;

  // Deployment to represent
  public deployment: Deployment;

  ngOnInit() {
    this.iconUrl = this.getIconUrl();
  }

  getIconUrl(): string {
    if (this.deployment.attributes.chartIcon) {
      return this.deployment.attributes.chartIcon;
    }
    return '/assets/images/placeholder.png';
  }
}
