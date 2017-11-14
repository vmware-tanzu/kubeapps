import { Observable } from 'rxjs/Rx';
import { ConfirmDialog } from '../../confirm-dialog/confirm-dialog.component';
import { MatDialogRef, MatDialog, MatDialogConfig } from '@angular/material';
import { Injectable } from '@angular/core';

@Injectable()
export class DialogsService {

    constructor(private dialog: MatDialog) { }

    public confirm(title: string, message: string, ok = 'Continue',
      cancel = 'Cancel', actionButtonClass = 'accent'): Observable<boolean> {
      let dialogRef: MatDialogRef<ConfirmDialog>;

      dialogRef = this.dialog.open(ConfirmDialog);
      dialogRef.componentInstance.title = title;
      dialogRef.componentInstance.message = message;
      dialogRef.componentInstance.actionButtonClass = actionButtonClass;
      dialogRef.componentInstance.ok = ok;
      dialogRef.componentInstance.cancel = cancel;

      return dialogRef.afterClosed();
    }
}
