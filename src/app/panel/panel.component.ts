import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-panel',
  templateUrl: './panel.component.html',
  styleUrls: ['./panel.component.scss'],
  inputs: ['title', 'background', 'container', 'border']
})
export class PanelComponent {
  // Title of the panel
  public title: string = '';
  // Display a gray background
  public background: boolean = false;
  // Show a border
  public border: boolean = false;
  // Set the size of the panel to 80%
  public container: boolean = false;
}
