////////////////////////////////////////////////////////////////////////////////////////////////////

use crate::cli::{CloneArgs, RememberArgs, StatsArgs, StatusArgs, SyncArgs};
use crate::util;
use anyhow::{Result, bail};
use std::io::Write;
use std::process::Command;

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run_clone(args: &CloneArgs) -> Result<()> {
    let target_dir = args.directory.as_deref().unwrap_or(".");
    util::clone_repositories_from_csv(&args.csv, target_dir)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run_describe(verbose: bool) -> Result<()> {
    let repos = util::collect_repos(None, verbose)?;
    util::print_describe_table(&repos, None)?;
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run_readme() -> Result<()> {
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
    let status = Command::new("bash").arg("-c").arg(cmd_str).status()?;
    if !status.success() {
        bail!("external command failed");
    }
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run_roadmap() -> Result<()> {
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
    let status = Command::new("bash").arg("-c").arg(cmd_str).status()?;
    if !status.success() {
        bail!("external command failed");
    }
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run_remember(args: RememberArgs, verbose: bool) -> Result<()> {
    let repos = util::collect_repos(None, verbose)?;
    let mut writer: Box<dyn Write> = if let Some(path) = &args.output {
        Box::new(std::fs::File::create(path)?)
    } else {
        Box::new(std::io::stdout())
    };
    util::write_remember_csv(&repos, &mut writer)?;
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run_stats(args: StatsArgs, verbose: bool) -> Result<()> {
    let specific = if args.repo == "." {
        None
    } else {
        Some(args.repo.clone())
    };
    let repos = util::collect_repos(specific, verbose)?;
    util::print_stats_table(&repos, args.year, &args.output)?;
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run_status(args: StatusArgs, verbose: bool) -> Result<()> {
    let repos = util::collect_repos(args.repo.clone(), verbose)?;
    let statuses = util::get_statuses(&repos, args.fetch)?;
    util::print_status_table(&statuses);
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn run_sync(args: SyncArgs, verbose: bool) -> Result<()> {
    let repos = util::collect_repos(args.repo.clone(), verbose)?;
    let action = if args.push { "push" } else { "pull" };
    let results = util::sync_repos(&repos, args.push, args.pull)?;
    util::print_sync_table(&results, action);
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////
