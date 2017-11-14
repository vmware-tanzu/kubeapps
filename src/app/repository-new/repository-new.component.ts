import { MatDialogRef } from '@angular/material';
import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { ReposService } from '../shared/services/repos.service';
import { Repo, RepoAttributes } from '../shared/models/repo';

@Component({
  selector: 'app-repository-new',
  templateUrl: './repository-new.component.html',
  styleUrls: ['./repository-new.component.scss'],
})
export class RepositoryNewComponent {
  repoAttributes: RepoAttributes = new RepoAttributes();
  formError: string;

  constructor(
    public dialogRef: MatDialogRef<RepositoryNewComponent>,
    private reposService: ReposService,
    private router: Router
  ) {}

  addRepo() {
    this.reposService.createRepo(this.repoAttributes)
      .subscribe(
        repo => {
          this.dialogRef.close();
        },
        error => {
          this.formError = error;
        }
      )
  }
}
