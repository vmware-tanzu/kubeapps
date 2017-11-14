import { MonocularPage } from './app.po';

describe('monocular App', () => {
  let page: MonocularPage;

  beforeEach(() => {
    page = new MonocularPage();
  });

  it('should display welcome message', () => {
    page.navigateTo();
    expect(page.getParagraphText()).toEqual('Welcome to app!');
  });
});
