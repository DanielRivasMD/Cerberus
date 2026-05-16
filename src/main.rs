////////////////////////////////////////////////////////////////////////////////////////////////////

mod cli;
mod cmd;
mod util;

////////////////////////////////////////////////////////////////////////////////////////////////////

use anyhow::Result;
use clap::FromArgMatches;

////////////////////////////////////////////////////////////////////////////////////////////////////

fn main() -> Result<()> {
    let mut cli_app = cli::build();
    let matches = cli_app.clone().get_matches();

    match matches.subcommand() {
        Some(("completion", sub_m)) => {
            let shell = sub_m.get_one::<clap_complete::Shell>("shell").unwrap();
            clap_complete::generate(*shell, &mut cli_app, "cerberus", &mut std::io::stdout());
        }
        Some(("identity", _)) => {
            println!("\n{}\n", cli::IDENT);
        }
        Some(("clone", sub_m)) => {
            let args = cli::CloneArgs::from_arg_matches(sub_m)?;
            cmd::run_clone(&args)?;
        }
        Some(("describe", _)) => {
            let verbose = matches.get_one::<bool>("verbose").copied().unwrap_or(false);
            cmd::run_describe(verbose)?;
        }
        Some(("readme", _)) => {
            cmd::run_readme()?;
        }
        Some(("remember", sub_m)) => {
            let args = cli::RememberArgs::from_arg_matches(sub_m)?;
            let verbose = matches.get_one::<bool>("verbose").copied().unwrap_or(false);
            cmd::run_remember(args, verbose)?;
        }
        Some(("roadmap", _)) => {
            cmd::run_roadmap()?;
        }
        Some(("stats", sub_m)) => {
            let args = cli::StatsArgs::from_arg_matches(sub_m)?;
            let verbose = matches.get_one::<bool>("verbose").copied().unwrap_or(false);
            cmd::run_stats(args, verbose)?;
        }
        Some(("status", sub_m)) => {
            let args = cli::StatusArgs::from_arg_matches(sub_m)?;
            let verbose = matches.get_one::<bool>("verbose").copied().unwrap_or(false);
            cmd::run_status(args, verbose)?;
        }
        Some(("sync", sub_m)) => {
            let args = cli::SyncArgs::from_arg_matches(sub_m)?;
            let verbose = matches.get_one::<bool>("verbose").copied().unwrap_or(false);
            cmd::run_sync(args, verbose)?;
        }
        _ => unreachable!(),
    }

    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////
