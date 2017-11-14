import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'truncate'
})
export class TruncatePipe implements PipeTransform {
  transform(value: string, exponent: string) : string {
    let limit = exponent ? parseInt(exponent, 80) : 80;

    return value.length > limit ? `${value.substring(0, limit - 3)}...` : value;
  }
}
