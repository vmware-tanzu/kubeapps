import { Injectable } from '@angular/core';
import { ConfigService } from './config.service';
import { MetaService } from '@ngx-meta/core';

// Import SEO data
import SeoData from '../seo.data';

/* TODO, This is a mocked class. */
@Injectable()
export class SeoService {
  constructor(private config: ConfigService, private metaService: MetaService) {}

  /**
   * Return the SEO data for the metaTags of the current page
   */
  getMetaContent(page, data = {}) {
    let metadata = Object.assign({}, SeoData[page]);
    // Regex of custom data
    let regex = /{ (\w+) }/i;
    let match;
    Object.keys(metadata).forEach(key => {
      while ((match = regex.exec(metadata[key]))) {
        metadata[key] = metadata[key].replace(match[0], data[match[1]]);
      }
    });

    return metadata;
  }

  /**
   * Set the given tags in the head of the page through MetaService
   */
  setMetaTags(page, data = {}) {
    let content = this.getMetaContent(page, data);
    // Set tags
    this.metaService.setTitle(content.title);
    this.metaService.setTag('description', content.description);
    this.metaService.setTag('og:title', content.title);
    this.metaService.setTag('og:description', content.description);

    // Check if we can set the image
    if (data['image'] !== undefined) {
      this.metaService.setTag('og:image', data['image']);
    }
  }
}
