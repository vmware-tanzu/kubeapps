import { Component, ViewEncapsulation } from '@angular/core';

@Component({
  selector: 'app-footer-list',
  templateUrl: './footer-list.component.html',
  styleUrls: ['./footer-list.component.scss'],
  inputs: ['title'],
  encapsulation: ViewEncapsulation.None
})
export class FooterListComponent {
  // Title of the panel
  public title:string = '';
}
