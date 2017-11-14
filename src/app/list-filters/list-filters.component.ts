import { Component } from '@angular/core';

@Component({
  selector: 'app-list-filters',
  templateUrl: './list-filters.component.html',
  styleUrls: ['./list-filters.component.scss'],
  inputs: ['filters']
})

export class ListFiltersComponent{
  public filters: { title: string, items: Array<{}> }[] = [];
}
