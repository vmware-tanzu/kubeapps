import { Component, OnInit, ViewEncapsulation, ChangeDetectionStrategy } from '@angular/core';

@Component({
  selector: 'app-deployment-resource',
  templateUrl: './deployment-resource.component.html',
  styleUrls: ['./deployment-resource.component.scss'],
  encapsulation: ViewEncapsulation.Emulated,
  changeDetection: ChangeDetectionStrategy.Default,
  inputs: ['resource']
})
export class DeploymentResourceComponent implements OnInit {
  // Resource to represent
  public resource;
  public serviceKeys: String[] = [];

  ngOnInit() {
    for(let k in this.resource.services[0]){
      this.serviceKeys.push(k);
    }
  }
}
