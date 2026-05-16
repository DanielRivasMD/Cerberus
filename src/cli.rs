////////////////////////////////////////////////////////////////////////////////////////////////////

use chrono::Datelike;
use clap::{Args, CommandFactory, FromArgMatches, Parser, Subcommand, ValueHint};
use clap_complete::Shell;

////////////////////////////////////////////////////////////////////////////////////////////////////

pub const IDENT: &str = r#"In Greek mythology, Cerberus, Κέρβερος, often referred to as the hound of Hades, is a multi-headed dog
that guards the gates of the underworld to prevent the dead from leaving.

He was the offspring of the monsters Echidna and Typhon, and was usually described as having three heads,
a serpent for a tail, and snakes protruding from his body.

Cerberus is primarily known for his capture by Heracles, the last of Heracles' twelve labours"#;

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Parser)]
#[command(
    name = "cerberus",
    version = "0.1.0",
    about = "Guardian of the code",
    long_about = None
)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Commands,

    /// Enable verbose diagnostics
    #[arg(global = true, short = 'v', long)]
    pub verbose: bool,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Subcommand)]
pub enum Commands {
    /// Generate shell completions
    Completion(CompletionArgs),

    /// Print the identity of Cerberus
    #[command(alias = "id")]
    Identity,

    /// Clone all repos from a CSV file
    Clone(CloneArgs),

    /// Explore repo descriptions
    Describe,

    /// Browse through README files
    Readme,

    /// Recall repos as CSV
    Remember(RememberArgs),

    /// Browse through roadmaps
    Roadmap,

    /// Report repos stats
    Stats(StatsArgs),

    /// Show status of multiple Git repositories
    Status(StatusArgs),

    /// Push or pull changes in Git repositories (only if clean)
    Sync(SyncArgs),
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Args)]
pub struct CompletionArgs {
    #[arg(value_enum)]
    pub shell: Shell,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Args)]
pub struct CloneArgs {
    /// File in csv format containing remote repositories
    #[arg(long, value_hint = ValueHint::FilePath)]
    pub csv: String,

    /// Directory to clone repositories into
    #[arg(long, value_hint = ValueHint::DirPath)]
    pub directory: Option<String>,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Args)]
pub struct RememberArgs {
    /// Output file (stdout if omitted)
    #[arg(long = "output", short = 'o', value_hint = ValueHint::FilePath)]
    pub output: Option<String>,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Args)]
pub struct StatsArgs {
    /// Repository path (default: current directory)
    #[arg(short, long, default_value = ".")]
    pub repo: String,

    /// Year for commit frequency calculation
    #[arg(short, long, default_value_t = chrono::Utc::now().year())]
    pub year: i32,

    /// Time aggregation (not yet implemented)
    #[arg(short, long, default_value = "yearly")]
    pub time: String,

    /// Render as graph (not yet implemented)
    #[arg(short, long, default_value_t = true)]
    pub plot: bool,

    /// Write output to CSV file instead of Markdown table
    #[arg(long, value_hint = ValueHint::FilePath)]
    pub output: Option<String>,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Args)]
pub struct StatusArgs {
    /// Specific repository path (default: scan subdirectories)
    #[arg(short, long)]
    pub repo: Option<String>,

    /// Run git fetch before checking upstream
    #[arg(short, long, default_value_t = false)]
    pub fetch: bool,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

#[derive(Args)]
#[group(required = true, multiple = false)]
pub struct SyncArgs {
    /// Specific repository path (default: scan subdirectories)
    #[arg(short, long)]
    pub repo: Option<String>,

    /// Push commits to remote
    #[arg(long)]
    pub push: bool,

    /// Pull commits from remote
    #[arg(long)]
    pub pull: bool,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn build() -> clap::Command {
    Cli::command()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
