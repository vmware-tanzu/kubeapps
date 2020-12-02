use structopt::StructOpt;

// Ensure the root crate is aware of the child modules.
mod cli;

fn main() {
    let opt = cli::Options::from_args();
    println!("{:#?}", opt);
}
