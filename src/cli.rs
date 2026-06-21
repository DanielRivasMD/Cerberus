////////////////////////////////////////////////////////////////////////////////////////////////////

use chrono::Datelike;
use clap::{Parser, Subcommand, ValueEnum, ValueHint};

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: add examples in help
const HELP: &str = r"";

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Parser)]
#[command(
    name = env!("CARGO_PKG_NAME"),
    version = env!("CARGO_PKG_VERSION"),
    author = env!("CARGO_PKG_AUTHORS"),
    about = env!("CARGO_PKG_DESCRIPTION"),
    before_help = concat!(env!("CARGO_PKG_AUTHORS"), "\n", env!("CARGO_PKG_NAME"), " v", env!("CARGO_PKG_VERSION")),
    long_about = HELP,
)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Command,

    /// Enable verbose diagnostics
    #[arg(global = true, short = 'v', long)]
    pub verbose: bool,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Subcommand)]
pub enum Command {
    /// Explore repos
    Explore {
        #[command(subcommand)]
        sub: ExploreSub,
    },

    /// Manage repos
    Manage {
        #[command(subcommand)]
        sub: ManageSub,
    },

    /// Print identity
    #[command(hide = true)]
    #[command(aliases = &["id"])]
    Identity,

    /// Generate shell completions
    #[command(hide = true)]
    Completion {
        /// Shell for which to generate completions
        #[arg(value_enum)]
        shell: Shell,
    },
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Subcommand)]
pub enum ExploreSub {
    /// Explore repo descriptions
    Describe,

    /// Browse through README files
    Readme,

    /// Browse through roadmaps
    Roadmap,

    /// Report repos stats
    Stats {
        /// Repository path (default: current directory)
        #[arg(short, long, default_value = ".")]
        repo: String,

        /// Year for commit frequency calculation
        #[arg(short, long, default_value_t = chrono::Utc::now().year())]
        year: i32,
        /// Time aggregation (not yet implemented)
        #[arg(short, long, default_value = "yearly")]
        time: String,

        /// Write output to CSV file instead of Markdown table
        #[arg(long, value_hint = ValueHint::FilePath)]
        output: Option<String>,
    },
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Subcommand)]
pub enum ManageSub {
    /// Clone all repos from a CSV file
    Clone {
        /// File in csv format containing remote repositories
        #[arg(long, value_hint = ValueHint::FilePath)]
        csv: String,

        /// Directory to clone repositories into
        #[arg(long, value_hint = ValueHint::DirPath)]
        directory: Option<String>,
    },

    /// Run git fetch before checking upstream
    Fetch {
        #[arg(short, long)]
        repo: Option<String>,
    },

    /// Run git pull
    Pull {
        /// Specific repository path (default: scan subdirectories)
        #[arg(short, long)]
        repo: Option<String>,
    },

    /// Run git push
    Push {
        /// Specific repository path (default: scan subdirectories)
        #[arg(short, long)]
        repo: Option<String>,
    },

    /// Recall repos as CSV
    Remember {
        /// File in csv format containing remote repositories
        #[arg(long, value_hint = ValueHint::FilePath)]
        csv: Option<String>,
    },

    /// Show status of multiple Git repositories
    Status {
        /// Specific repository path (default: scan subdirectories)
        #[arg(short, long)]
        repo: Option<String>,
    },
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Clone, Copy, ValueEnum)]
pub enum Shell {
    Bash,
    Zsh,
    Fish,
    Powershell,
}

////////////////////////////////////////////////////////////////////////////////////////////////////
