import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-list-item',
  templateUrl: './list-item.component.html',
  styleUrls: ['./list-item.component.scss'],
  inputs: ['detailUrl']
})
export class ListItemComponent {
  public detailUrl: string;
}
