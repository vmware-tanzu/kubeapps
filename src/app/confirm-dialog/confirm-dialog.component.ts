import { MatDialogRef } from '@angular/material';
import { Component } from '@angular/core';

@Component({
    selector: 'confirm-dialog',
    templateUrl: './confirm-dialog.component.html',
    styleUrls: ['./confirm-dialog.component.scss'],
})
export class ConfirmDialog {

    public title: string;
    public message: string;
    public ok: string;
    public cancel: string;
    public actionButtonClass: string;

    constructor(public dialogRef: MatDialogRef<ConfirmDialog>) {

    }
}
