////////////////////////////////////////////////////////////////////////////////////////////////////

use anyhow::{Result as anyResult, bail};
use std::process::Command;

////////////////////////////////////////////////////////////////////////////////////////////////////

use crate::cli;
use crate::util;

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run(sub: cli::ExploreSub, verbose: bool) -> anyResult<()> {
    match sub {
        cli::ExploreSub::Describe => describe::run(verbose)?,
        cli::ExploreSub::Readme => readme::run()?,
        cli::ExploreSub::Roadmap => roadmap::run()?,
        cli::ExploreSub::Stats {
            repo,
            year,
            time,
            output,
        } => stats::run(repo, year, time, output, verbose)?,
    }
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: describe does not use remote URL
mod describe {
    pub fn run(verbose: bool) -> super::anyResult<()> {
        let repos = super::util::collect_repos(None, verbose)?;
        super::util::print_describe_table(&repos, None)?;
        Ok(())
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

mod readme {
    pub fn run() -> super::anyResult<()> {
        let cmd_str = r#"
zellij run --name readme \
    --close-on-exit --floating \
    --height 100 --width 130 --x 15 --y 0 \
    -- zsh -c '
        file=$(
            fd --type f --glob "README.md" . \
            | fzf \
                --preview="bat --style=plain --color=always {}" \
                --preview-window="right:70%" \
                --height=100% \
                --reverse
        )
    [[ -n $file ]] && mdcat --paginate --columns=100 "$file"
    '"#;
        let status = super::Command::new("bash")
            .arg("-c")
            .arg(cmd_str)
            .status()?;
        if !status.success() {
            super::bail!("external command failed");
        }
        Ok(())
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

mod roadmap {
    pub fn run() -> super::anyResult<()> {
        let cmd_str = r#"
zellij run --name roadmap \
	--close-on-exit --floating \
	--height 100 --width 130 --x 15 --y 0 \
	-- zsh -c '
		file=$(
			fd --type f --glob 'ROADMAP.txt' . \
			| fzf \
				--preview="bat --style=plain --color=always {}" \
				--preview-window="right:70%" \
				--height=100% \
				--reverse
		)
	[[ -n $file ]] && hx "$file"
    '"#;
        let status = super::Command::new("bash")
            .arg("-c")
            .arg(cmd_str)
            .status()?;
        if !status.success() {
            super::bail!("external command failed");
        }
        Ok(())
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: command aggregation on quatterly
mod stats {
    pub fn run(
        repo: String,
        year: i32,
        time: String,
        output: Option<String>,
        verbose: bool,
    ) -> super::anyResult<()> {
        let specific = if repo == "." {
            None
        } else {
            Some(repo.clone())
        };
        let repos = super::util::collect_repos(specific, verbose)?;
        super::util::print_stats_table(&repos, year, &output)?;
        Ok(())
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////
