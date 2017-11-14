import { Component } from '@angular/core';

@Component({
  selector: 'app-loader',
  templateUrl: './loader.component.html',
  styleUrls: ['./loader.component.scss'],
  inputs: ['loading']
})
export class LoaderComponent {
  // Show the loader or the content
  public loading: boolean = false;

  constructor() {}
}
