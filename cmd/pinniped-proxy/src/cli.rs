use structopt::StructOpt;

#[derive(StructOpt)]
/// Converts requests with bearer tokens to requests with short-lived certs.
///
/// pinniped-proxy proxies incoming requests with an `Authorization: Bearer
/// token` header upstream as requests with short-lived client certificates,
/// where the bearer token has been exchanged for the client certs using the
/// pinniped aggregate API.
pub struct Options {
    #[structopt(
        short = "p",
        long = "port",
        default_value = "3333",
        help = "Specify the port on which pinniped-proxy listens."
    )]
    pub port: u16, 

    #[structopt(
        long = "pinniped-executable",
        short = "x",
        default_value = "pinniped",
        help = "The name of the executable, including the full path if required",
    )]
    pub pinniped_executable: String,
}
