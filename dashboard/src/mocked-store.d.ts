// TODO(miguel) I have not managed to make the linter work and also
// enhance the global JSX namespace without disabling it.
/* tslint:disable */
declare namespace JSX {
  interface IntrinsicAttributes {
    // Add store option for testing with mocked-store
    store?: any;
  }
}
