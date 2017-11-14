import { Component, OnInit } from '@angular/core';
import { DeploymentsService } from '../shared/services/deployments.service';
import { Deployment } from '../shared/models/deployment';
import { Router } from '@angular/router';
import { SeoService } from '../shared/services/seo.service';
import { ConfigService } from '../shared/services/config.service';
import { DomSanitizer } from '@angular/platform-browser';
import { MatIconRegistry } from '@angular/material';

@Component({
  selector: 'app-deployments',
  templateUrl: './deployments.component.html',
  styleUrls: ['./deployments.component.scss'],
  viewProviders: [MatIconRegistry]
})
export class DeploymentsComponent implements OnInit {
  deployments: Deployment[] = [];
  visibleDeployments: Deployment[] = [];
  loading: boolean = true;
  filtersOpen: boolean = false;
  orderBy: string = 'Date';
  namespace: string = 'All';
  filters: Array<any> = [
    {
      title: 'Namespace',
      onSelect: i => this.onSelectNamespace(i),
      items: [{ title: 'All', selected: true }]
    },
    {
      title: 'Order By',
      onSelect: i => this.onSelectOrderBy(i),
      items: [
        { title: 'Name', selected: false },
        { title: 'Date', selected: true },
        { title: 'Status', selected: false }
      ]
    }
  ];

  constructor(
    private deploymentsService: DeploymentsService,
    private router: Router,
    private seo: SeoService,
    private config: ConfigService,
    private mdIconRegistry: MatIconRegistry,
    private sanitizer: DomSanitizer
  ) {}

  ngOnInit() {
    this.mdIconRegistry.addSvgIcon(
      'search',
      this.sanitizer.bypassSecurityTrustResourceUrl(`/assets/icons/search.svg`)
    );
    this.mdIconRegistry.addSvgIcon(
      'close',
      this.sanitizer.bypassSecurityTrustResourceUrl(`/assets/icons/close.svg`)
    );
    this.mdIconRegistry.addSvgIcon(
      'menu',
      this.sanitizer.bypassSecurityTrustResourceUrl(`/assets/icons/menu.svg`)
    );
    // Do not show the page if the feature is not enabled
    if (!this.config.releasesEnabled) {
      return this.router.navigate(['/404']);
    }
    this.seo.setMetaTags('deployments');
    this.loadDeployments();
  }

  loadDeployments(): void {
    this.deploymentsService
      .getDeployments()
      .finally(() => {
        this.loading = false;
      })
      .subscribe(deployments => {
        this.deployments = deployments;
        this.filterDeployments();
        this.exportNamespaces();
      });
  }

  exportNamespaces() {
    var flags = {};
    var list: { title: string; selected: boolean }[] = [
      { title: 'All', selected: true }
    ];
    this.deployments.forEach(dp => {
      if (!flags[dp.attributes.namespace]) {
        list.push({
          title: dp.attributes.namespace,
          selected: false
        });
        flags[dp.attributes.namespace] = true;
      }
    });
    this.filters[0].items = list;
  }

  filterDeployments() {
    let filtered = this.deployments;
    if (this.namespace !== 'All') {
      filtered = filtered.filter(deployment => {
        return deployment.attributes.namespace === this.namespace;
      });
    }
    filtered = filtered.sort((a, b) => {
      if (this.orderBy === 'Name') {
        return a.id <= b.id ? -1 : 1;
      } else if (this.orderBy === 'Status') {
        return a.attributes.status <= b.attributes.status ? -1 : 1;
      } else {
        return a.attributes.updated <= b.attributes.updated ? -1 : 1;
      }
    });
    this.visibleDeployments = filtered;
  }

  searchChange(e) {
    let newValue = e.target.value;
    if (!newValue) {
      return this.filterDeployments();
    }
    let searchTerm = newValue.toLowerCase();
    this.visibleDeployments = this.deployments.filter(deployment => {
      return deployment.id.indexOf(searchTerm) != -1;
    });
  }

  onSelectNamespace(index) {
    this.namespace = this.filters[0].items[index].title;
    this.filters[0].items = this.filters[0].items.map(n => {
      n.selected = n.title == this.namespace;
      return n;
    });
    this.filterDeployments();
  }

  onSelectOrderBy(index) {
    this.orderBy = this.filters[1].items[index].title;
    this.filters[1].items = this.filters[1].items.map(o => {
      o.selected = o.title == this.orderBy;
      return o;
    });
    this.filterDeployments();
  }
}
