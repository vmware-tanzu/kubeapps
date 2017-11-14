import {
  Component,
  OnInit,
  Input,
  ViewEncapsulation,
  ChangeDetectionStrategy
} from '@angular/core';
import { Chart } from '../../shared/models/chart';
import { Deployment } from '../../shared/models/deployment';
import { DomSanitizer } from '@angular/platform-browser';
import { MatIconRegistry, MatSnackBar, MatDialog, MatDialogRef } from '@angular/material';
import { ConfigService } from '../../shared/services/config.service';
import { DeploymentsService } from '../../shared/services/deployments.service';
import { DeploymentNewComponent } from '../../deployment-new/deployment-new.component';
import { Router } from '@angular/router';

@Component({
  selector: 'app-chart-details-usage',
  templateUrl: './chart-details-usage.component.html',
  styleUrls: ['./chart-details-usage.component.scss'],
  viewProviders: [MatIconRegistry],
  encapsulation: ViewEncapsulation.None
})
export class ChartDetailsUsageComponent implements OnInit {
  @Input() chart: Chart;
  @Input() currentVersion: string;
  installing: boolean;

  constructor(
    private mdIconRegistry: MatIconRegistry,
    private sanitizer: DomSanitizer,
    public config: ConfigService,
    private deploymentsService: DeploymentsService,
    private router: Router,
    private dialog: MatDialog,
    public snackBar: MatSnackBar
  ) {}

  ngOnInit() {
    this.mdIconRegistry.addSvgIcon(
      'content-copy',
      this.sanitizer.bypassSecurityTrustResourceUrl(
        '/assets/icons/content-copy.svg'
      )
    );
  }

  // Show an snack bar to confirm the user that the code has been copied
  showSnackBar(): void {
    this.snackBar.open('Copied to the clipboard', '', {
      duration: 1500
    });
  }

  get showRepoInstructions(): boolean {
    return this.chart.attributes.repo.name != 'stable';
  }

  get repoAddInstructions(): string {
    return `helm repo add ${this.chart.attributes.repo.name} ${this.chart
      .attributes.repo.URL}`;
  }

  get installInstructions(): string {
    return `helm install ${this.chart.id} --version ${this.currentVersion}`;
  }

  installDeployment(chartID: string, version: string): void {
    let dialogRef: MatDialogRef<DeploymentNewComponent>;
    dialogRef = this.dialog.open(DeploymentNewComponent);
    dialogRef.afterClosed()
      .subscribe(namespace => {
        if (namespace) this.performInstallation(chartID, version, namespace);
      });
    dialogRef.componentInstance.chartID = chartID;
    dialogRef.componentInstance.version = version;
  }

  performInstallation(chartID: string, version: string, namespace: string): void {
    this.installing = true;

    this.deploymentsService
      .installDeployment(chartID, version, namespace)
      .finally(() => {
        this.installing = false;
      })
      .subscribe(
        deployment => {
          this.installOK(deployment);
        },
        error => {
          this.snackBar.open(
            `Error installing the application, please try later`,
            'close',
            { duration: 5000 }
          );
        }
      );
  }

  installOK(deployment: Deployment): void {
    this.router.navigate(['/deployments', deployment.id]);
  }
}
