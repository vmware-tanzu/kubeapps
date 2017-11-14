import { Component, Output, EventEmitter } from '@angular/core';
import { DeploymentsService } from '../shared/services/deployments.service';
import { MatSnackBar } from '@angular/material';
import { DialogsService } from '../shared/services/dialogs.service';
import { Deployment } from '../shared/models/deployment';

@Component({
  selector: 'app-deployment-controls',
  templateUrl: './deployment-controls.component.html',
  styleUrls: ['./deployment-controls.component.scss'],
  inputs: ['deployment']
})
export class DeploymentControlsComponent {
  public deployment: Deployment;
  @Output() onDelete = new EventEmitter();
  deleting: boolean

  constructor(
    private deploymentsService: DeploymentsService,
    private dialogsService: DialogsService,
    public snackBar: MatSnackBar
  ){ }

  deleteDeployment(deploymentName: string): void {
    this.dialogsService
      .confirm(
        `Are you sure you want to delete "${deploymentName}"?`,
        'You will need to launch a new deployment',
        `Delete ${deploymentName}`,
        'Cancel',
        'warn'
      )
      .subscribe(res => {
        if(res) {
          this.performDelete(deploymentName);
        }
      })
  }

  performDelete(deploymentName: string): void {
    this.deleting = true;
    this.onDelete.emit({ name: deploymentName, state: "deleting" });
    this.snackBar.open("Deleting deployment", "close", {});
    this.deploymentsService.deleteDeployment(deploymentName)
    .finally(() => {
      this.deleting = false
    }).subscribe(
      deployment => {
        this.onDelete.emit({ name: deploymentName, state: "deleted" });
        this.snackBar.open("Deployment deleted", "", { duration: 2500 });
      },
      error => {
        this.snackBar.open("Error deleting the deployment", "", { duration: 2500 });
      }
    );
  }
}
