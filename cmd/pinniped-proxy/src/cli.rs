use structopt::StructOpt;

#[derive(StructOpt, Debug)]
/// A proxy server which converts k8s API server requests with bearer tokens to
/// requests with short-lived X509 certs exchanged by pinniped.
///
/// pinniped-proxy proxies incoming requests with an `Authorization: Bearer
/// token` header, exchanging the token via the pinniped aggregate API for x509
/// short-lived client certificates, before forwarding the request onwards
/// to the destination k8s API server.
pub struct Options {
    #[structopt(
        short = "p",
        long = "port",
        default_value = "3333",
        help = "Specify the port on which pinniped-proxy listens."
    )]
    pub port: u16, 
}
