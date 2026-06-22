////////////////////////////////////////////////////////////////////////////////////////////////////

use anyhow::Result as anyResult;
use std::io::Write;

////////////////////////////////////////////////////////////////////////////////////////////////////

use crate::cli;
use crate::util;

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run(sub: cli::ManageSub, recursive: bool, verbose: bool) -> anyResult<()> {
    match sub {
        cli::ManageSub::Clone { csv, directory } => clone::run(csv, directory, verbose)?,
        cli::ManageSub::Fetch { repo } => fetch::run(repo, recursive, verbose)?,
        cli::ManageSub::Pull { repo } => pull::run(repo, recursive, verbose)?,
        cli::ManageSub::Push { repo } => push::run(repo, recursive, verbose)?,
        cli::ManageSub::Remember { csv } => remember::run(csv, recursive, verbose)?,
        cli::ManageSub::Status { repo } => status::run(repo, recursive, verbose)?,
    }
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

mod clone {
    pub fn run(csv: String, directory: Option<String>, verbose: bool) -> super::anyResult<()> {
        let target_dir = directory.as_deref().unwrap_or(".");
        super::util::clone_repositories_from_csv(&csv, target_dir)
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

mod fetch {
    pub fn run(repo: Option<String>, recursive: bool, verbose: bool) -> super::anyResult<()> {
        super::util::status_report(repo, true, recursive, verbose)
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

mod remember {
    pub fn run(csv: Option<String>, recursive: bool, verbose: bool) -> super::anyResult<()> {
        let repos = super::util::collect_repos(None, recursive, verbose)?;
        let mut writer: Box<dyn super::Write> = if let Some(path) = &csv {
            Box::new(std::fs::File::create(path)?)
        } else {
            Box::new(std::io::stdout())
        };
        super::util::write_remember_csv(&repos, &mut writer)?;
        Ok(())
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

mod status {
    pub fn run(repo: Option<String>, recursive: bool, verbose: bool) -> super::anyResult<()> {
        super::util::status_report(repo, false, recursive, verbose)
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

mod pull {
    pub fn run(repo: Option<String>, recursive: bool, verbose: bool) -> super::anyResult<()> {
        let repos = super::util::collect_repos(repo.clone(), recursive, verbose)?;
        let action = "pull";
        let results = super::util::sync_repos(&repos, false, true)?;
        super::util::print_sync_table(&results, action);
        Ok(())
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

mod push {
    pub fn run(repo: Option<String>, recursive: bool, verbose: bool) -> super::anyResult<()> {
        let repos = super::util::collect_repos(repo.clone(), recursive, verbose)?;
        let action = "push";
        let results = super::util::sync_repos(&repos, true, false)?;
        super::util::print_sync_table(&results, action);
        Ok(())
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////
