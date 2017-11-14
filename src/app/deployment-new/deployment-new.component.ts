import { MatDialogRef } from '@angular/material';
import { Component } from '@angular/core';

@Component({
  selector: 'app-deployment-new',
  templateUrl: './deployment-new.component.html',
  styleUrls: ['./deployment-new.component.scss'],
})
export class DeploymentNewComponent {
  public chartID: string;
  public version: string;
  public namespace: string = 'default';

  constructor(
    public dialogRef: MatDialogRef<DeploymentNewComponent>,
  ) {}
}
