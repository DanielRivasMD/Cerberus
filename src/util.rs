////////////////////////////////////////////////////////////////////////////////////////////////////

use anyhow::{Context, Result as anyResult, bail};
use chrono::{Datelike, NaiveDateTime, Utc};
use colored::*;
use rayon::prelude::*;
use serde_json::Value;
use std::collections::HashMap;
use std::io::Write;
use std::path::Path;
use std::process::Command;
use std::sync::Mutex;
use walkdir::WalkDir;

////////////////////////////////////////////////////////////////////////////////////////////////////

trait Named {
    fn name(&self) -> &str;
}

impl Named for RepoStatus {
    fn name(&self) -> &str {
        &self.name
    }
}
impl Named for SyncResult {
    fn name(&self) -> &str {
        &self.name
    }
}
impl Named for RepoStats {
    fn name(&self) -> &str {
        &self.repo
    }
}
impl Named for RepoDescribe {
    fn name(&self) -> &str {
        &self.repo
    }
}
impl Named for (String, String) {
    fn name(&self) -> &str {
        &self.0
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

/// Process repos in parallel. The closure `f` receives a repo path and returns a `anyResult<T>`.
/// Successful results are collected and sorted by repo name. Any errors are printed as warnings.
/// Returns a `Vec<T>` containing all successful results.
fn par_process_repos<T: Send + Named>(
    repos: &[String],
    f: impl Fn(&str) -> anyResult<T> + Sync,
) -> Vec<T>
where
    T: Send,
{
    let results = Mutex::new(Vec::new());
    repos.par_iter().for_each(|repo| match f(repo) {
        Ok(t) => {
            results.lock().unwrap().push(t);
        }
        Err(e) => {
            let name = Path::new(repo)
                .file_name()
                .unwrap_or(repo.as_ref())
                .to_string_lossy();
            eprintln!("Warning: {} (repo: {})", e, name);
        }
    });
    let mut results = results.into_inner().unwrap();
    results.sort_by_key(|item| item.name().to_lowercase());
    results
}

////////////////////////////////////////////////////////////////////////////////////////////////////

/// Collects repository paths (absolute). If `recursive` is true, also scan one level deep
/// inside each found repo for nested Git repositories.
pub fn collect_repos(
    specific: Option<String>,
    recursive: bool,
    verbose: bool,
) -> anyResult<Vec<String>> {
    let mut repos = if let Some(p) = specific {
        let abs = std::fs::canonicalize(&p).with_context(|| format!("resolving path: {}", p))?;
        if !abs.join(".git").is_dir() {
            bail!("{} is not a Git repository", abs.display());
        }
        vec![abs.to_string_lossy().into_owned()]
    } else {
        let cwd = std::env::current_dir()?;
        if cwd.join(".git").is_dir() {
            vec![cwd.to_string_lossy().into_owned()]
        } else {
            let mut v = Vec::new();
            for entry in std::fs::read_dir(&cwd)? {
                let entry = entry?;
                let path = entry.path();
                if path.is_dir() && path.join(".git").is_dir() {
                    v.push(path.to_string_lossy().into_owned());
                }
            }
            v
        }
    };

    if recursive {
        let mut nested = Vec::new();
        for repo in &repos {
            if let Ok(entries) = std::fs::read_dir(repo) {
                for entry in entries.flatten() {
                    let path = entry.path();
                    if path.is_dir() && path.join(".git").is_dir() {
                        // Avoid duplicates if nested repo was already listed
                        let nested_path = path.to_string_lossy().into_owned();
                        if !repos.contains(&nested_path) && !nested.contains(&nested_path) {
                            nested.push(nested_path);
                        }
                    }
                }
            }
        }
        repos.extend(nested);
    }

    Ok(repos)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

/// Runs git command with -C repo_path and returns trimmed stdout.
pub fn call_git(repo_path: &str, args: &[&str]) -> anyResult<String> {
    let mut cmd = Command::new("git");
    cmd.arg("-C").arg(repo_path);
    cmd.args(args);
    let output = cmd
        .output()
        .with_context(|| format!("running git {:?} in {}", args, repo_path))?;
    if !output.status.success() {
        bail!(
            "git {:?} failed: {}",
            args,
            String::from_utf8_lossy(&output.stderr)
        );
    }
    Ok(String::from_utf8_lossy(&output.stdout).trim().to_string())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

/// Clones repositories from CSV file.
pub fn clone_repositories_from_csv(csv_path: &str, target_dir: &str, dry_run: bool) -> anyResult<()> {
    let mut reader = csv::Reader::from_path(csv_path)?;
    let mut first = true;
    for result in reader.records() {
        let record = result?;
        if first && record.get(0).map(|s| s.to_lowercase()) == Some("reponame".to_string()) {
            first = false;
            continue;
        }
        let repo_name = record
            .get(0)
            .map(|s| s.trim().to_string())
            .unwrap_or_default();
        let repo_url = record
            .get(1)
            .map(|s| s.trim().to_string())
            .unwrap_or_default();
        if repo_name.is_empty() || repo_url.is_empty() {
            continue;
        }
        let dest = Path::new(target_dir).join(&repo_name);
        if dry_run {
            println!("[DRY RUN] Would clone {} into {}", repo_url, dest.display());
        } else {
            println!("Cloning {} into {}", repo_url, dest.display());
            let status = Command::new("git")
                .arg("clone")
                .arg(&repo_url)
                .arg(&dest)
                .status()?;
            if !status.success() {
                bail!("failed to clone {}", repo_url);
            }
        }
    }
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Describe
////////////////////////////////////////////////////////////////////////////////////////////////////

pub struct RepoDescribe {
    pub repo: String,
    pub overview: String,
    pub license: String,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn print_describe_table(repos: &[String], _output_file: Option<&str>) -> anyResult<()> {
    let rows = par_process_repos(repos, |repo| {
        let desc = populate_describe(repo)?;
        let name = Path::new(repo)
            .file_name()
            .unwrap()
            .to_string_lossy()
            .into_owned();
        Ok(RepoDescribe {
            repo: name,
            overview: truncate(&desc.overview, 92),
            license: desc.license,
        })
    });

    if rows.is_empty() {
        println!("No descriptions available.");
        return Ok(());
    }

    let widths = [25, 92, 7];
    let headers = vec!["Repo", "Overview", "License"];
    let align_left = [true, true, true];
    print_markdown_table(&headers, &widths, &align_left, &rows, |row| {
        vec![row.repo.clone(), row.overview.clone(), row.license.clone()]
    });
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn populate_describe(repo_path: &str) -> anyResult<RepoDescribe> {
    let base = Path::new(repo_path);
    let overview = if base.join("README.md").exists() {
        parse_readme(base.join("README.md").to_str().unwrap(), 92)?
    } else {
        String::new()
    };
    let license = if base.join("LICENSE").exists() {
        detect_license(base.join("LICENSE").to_str().unwrap())?
    } else {
        "Unknown".to_string()
    };
    Ok(RepoDescribe {
        repo: String::new(),
        overview,
        license,
    })
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn parse_readme(path: &str, max_chars: usize) -> anyResult<String> {
    let content = std::fs::read_to_string(path)?;
    let mut in_overview = false;
    let mut lines = Vec::new();
    for line in content.lines() {
        if line.trim_start().starts_with("## Overview") {
            in_overview = true;
            continue;
        }
        if in_overview && line.starts_with("## ") {
            break;
        }
        if in_overview {
            lines.push(line.trim().to_string());
        }
    }
    let result = lines.join("\n");
    let result = trim_to_period_or_newline(&result);
    Ok(truncate(&result, max_chars).to_string())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn detect_license(path: &str) -> anyResult<String> {
    let content = std::fs::read_to_string(path)?.to_lowercase();
    let keywords = [
        ("mit license", "MIT"),
        ("apache license", "Apache-2.0"),
        ("gnu general public license", "GPL"),
        ("bsd license", "BSD"),
        ("mozilla public license", "MPL"),
        ("creative commons", "CC"),
        ("eclipse public license", "EPL"),
    ];
    for (kw, lic) in keywords {
        if content.contains(kw) {
            return Ok(lic.to_string());
        }
    }
    Ok("Unknown".to_string())
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Remember
////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn write_remember_csv(repos: &[String], writer: &mut dyn Write) -> anyResult<()> {
    let entries = par_process_repos(repos, |repo| {
        let name = Path::new(repo)
            .file_name()
            .unwrap()
            .to_string_lossy()
            .into_owned();
        let remote = get_remote_url(repo).unwrap_or_default();
        Ok((name, remote))
    });

    let mut wtr = csv::Writer::from_writer(writer);
    wtr.write_record(["repoName", "repoURL"])?;
    for (name, remote) in entries {
        wtr.write_record(&[name, remote])?;
    }
    wtr.flush()?;
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn get_remote_url(repo_path: &str) -> anyResult<String> {
    call_git(repo_path, &["config", "--get", "remote.origin.url"])
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Stats
////////////////////////////////////////////////////////////////////////////////////////////////////

pub struct RepoStats {
    pub repo: String,
    pub commit: usize,
    pub age: String,
    pub language: String,
    pub lines: usize,
    pub size: String,
    pub mean: f64,
    pub q1: usize,
    pub q2: usize,
    pub q3: usize,
    pub q4: usize,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn print_stats_table(
    repos: &[String],
    year: i32,
    _output_file: &Option<String>,
) -> anyResult<()> {
    let rows = par_process_repos(repos, |repo| populate_repo_stats(repo, year));

    if rows.is_empty() {
        println!("No stats available.");
        return Ok(());
    }

    let widths = [25, 6, 6, 15, 6, 7, 4, 3, 3, 3, 3];
    let headers = vec![
        "Repo", "Commit", "Age", "Language", "Lines", "Size", "Mean", "Q1", "Q2", "Q3", "Q4",
    ];
    let align_left = [
        true, false, false, true, false, false, false, false, false, false, false,
    ];
    print_markdown_table(&headers, &widths, &align_left, &rows, |row| {
        vec![
            row.repo.clone(),
            row.commit.to_string(),
            row.age.clone(),
            row.language.clone(),
            row.lines.to_string(),
            row.size.clone(),
            format!("{:.1}", row.mean),
            row.q1.to_string(),
            row.q2.to_string(),
            row.q3.to_string(),
            row.q4.to_string(),
        ]
    });
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn populate_repo_stats(repo_path: &str, year: i32) -> anyResult<RepoStats> {
    let name = Path::new(repo_path)
        .file_name()
        .unwrap()
        .to_string_lossy()
        .into_owned();

    let tokei_json = call_tokei(repo_path)?;
    let (language, lines_percent, total_lines) = parse_tokei_json(&tokei_json)?;

    let age = repo_age(repo_path)?;
    let commit_count = count_commits(repo_path)?;
    let size = repo_size(repo_path)?;
    let freq = commit_frequency(repo_path, year)?;

    let age_months = parse_age_months(&age);
    let mean = if age_months > 0 {
        commit_count as f64 / age_months as f64
    } else {
        0.0
    };

    let q1 = freq.get(&format!("{}-01", year)).cloned().unwrap_or(0)
        + freq.get(&format!("{}-02", year)).cloned().unwrap_or(0)
        + freq.get(&format!("{}-03", year)).cloned().unwrap_or(0);
    let q2 = freq.get(&format!("{}-04", year)).cloned().unwrap_or(0)
        + freq.get(&format!("{}-05", year)).cloned().unwrap_or(0)
        + freq.get(&format!("{}-06", year)).cloned().unwrap_or(0);
    let q3 = freq.get(&format!("{}-07", year)).cloned().unwrap_or(0)
        + freq.get(&format!("{}-08", year)).cloned().unwrap_or(0)
        + freq.get(&format!("{}-09", year)).cloned().unwrap_or(0);
    let q4 = freq.get(&format!("{}-10", year)).cloned().unwrap_or(0)
        + freq.get(&format!("{}-11", year)).cloned().unwrap_or(0)
        + freq.get(&format!("{}-12", year)).cloned().unwrap_or(0);

    Ok(RepoStats {
        repo: name,
        commit: commit_count,
        age,
        language: format!("{} {}%", language, lines_percent),
        lines: total_lines,
        size,
        mean,
        q1,
        q2,
        q3,
        q4,
    })
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn call_tokei(repo_path: &str) -> anyResult<String> {
    let output = Command::new("tokei")
        .args(["-C", "-o", "json"]) // JSON output
        .current_dir(repo_path)
        .output()?;
    if !output.status.success() {
        bail!("tokei failed");
    }
    Ok(String::from_utf8_lossy(&output.stdout).trim().to_string())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn parse_tokei_json(json_str: &str) -> anyResult<(String, usize, usize)> {
    let v: Value = serde_json::from_str(json_str)?;
    let obj = v.as_object().context("tokei JSON is not an object")?;

    let mut total_lines = 0usize;
    let mut dominant_lang = String::new();
    let mut max_lines = 0u64;

    for (lang, entry) in obj {
        if lang.eq_ignore_ascii_case("Total") {
            // Compute total lines from the Total entry
            let code = entry["code"].as_u64().unwrap_or(0);
            let comments = entry["comments"].as_u64().unwrap_or(0);
            let blanks = entry["blanks"].as_u64().unwrap_or(0);
            total_lines = (code + comments + blanks) as usize;
            continue;
        }

        let code = entry["code"].as_u64().unwrap_or(0);
        let comments = entry["comments"].as_u64().unwrap_or(0);
        let blanks = entry["blanks"].as_u64().unwrap_or(0);
        let lines = code + comments + blanks;

        if lines > max_lines {
            max_lines = lines;
            dominant_lang = lang.clone();
        }
    }

    if dominant_lang.is_empty() {
        bail!("no language entries found in tokei output");
    }

    let pct = if total_lines > 0 {
        (max_lines as usize * 100) / total_lines
    } else {
        0
    };

    Ok((dominant_lang, pct, total_lines))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn repo_age(repo_path: &str) -> anyResult<String> {
    let out = call_git(repo_path, &["log", "--reverse", "--format=%ci"])?;
    let first = out.lines().next().unwrap_or("");
    if first.is_empty() {
        return Ok("0y 0m".to_string());
    }
    let naive = NaiveDateTime::parse_from_str(first, "%Y-%m-%d %H:%M:%S %z")
        .or_else(|_| NaiveDateTime::parse_from_str(first, "%Y-%m-%d %H:%M:%S"))?;
    let now = Utc::now().naive_utc();
    let diff = now.signed_duration_since(naive);
    let years = diff.num_days() / 365;
    let months = (diff.num_days() % 365) / 30;
    Ok(format!("{}y {}m", years, months))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn parse_age_months(age: &str) -> usize {
    let parts: Vec<&str> = age.split_whitespace().collect();
    let mut years = 0;
    let mut months = 0;
    for p in parts {
        if p.ends_with('y') {
            if let Ok(y) = p.trim_end_matches('y').parse::<usize>() {
                years = y;
            }
        } else if p.ends_with('m') {
            if let Ok(m) = p.trim_end_matches('m').parse::<usize>() {
                months = m;
            }
        }
    }
    years * 12 + months
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn count_commits(repo_path: &str) -> anyResult<usize> {
    let out = call_git(repo_path, &["rev-list", "--count", "HEAD"])?;
    Ok(out.parse::<usize>().unwrap_or(0))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn repo_size(repo_path: &str) -> anyResult<String> {
    let mut size: u64 = 0;
    for entry in WalkDir::new(repo_path).into_iter().filter_map(|e| e.ok()) {
        if entry.file_type().is_file() {
            size += entry.metadata().map(|m| m.len()).unwrap_or(0);
        }
    }
    Ok(format_size(size))
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn format_size(bytes: u64) -> String {
    const KB: u64 = 1024;
    const MB: u64 = KB * 1024;
    const GB: u64 = MB * 1024;
    if bytes >= GB {
        format!("{} GB", bytes / GB)
    } else if bytes >= MB {
        format!("{} MB", bytes / MB)
    } else if bytes >= KB {
        format!("{} KB", bytes / KB)
    } else {
        format!("{} bytes", bytes)
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn commit_frequency(repo_path: &str, year: i32) -> anyResult<HashMap<String, usize>> {
    let mut freq = HashMap::new();
    for m in 1..=12 {
        freq.insert(format!("{}-{:02}", year, m), 0);
    }
    let out = call_git(
        repo_path,
        &[
            "log",
            "--since",
            &format!("{}-01-01", year),
            "--until",
            &format!("{}-12-31", year),
            "--format=%ci",
        ],
    )?;
    for line in out.lines() {
        if let Ok(dt) = NaiveDateTime::parse_from_str(line, "%Y-%m-%d %H:%M:%S %z") {
            let key = format!("{}-{:02}", dt.year(), dt.month());
            *freq.entry(key).or_default() += 1;
        }
    }
    Ok(freq)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Status
////////////////////////////////////////////////////////////////////////////////////////////////////

pub struct RepoStatus {
    pub name: String,
    pub clean: bool,
    pub upstream: String,
    pub ahead: usize,
    pub behind: usize,
    pub error: Option<String>,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

/// Run a full status report: collect repos, optionally fetch, and print the table
pub fn status_report(
    repo: Option<String>,
    fetch: bool,
    recursive: bool,
    verbose: bool,
) -> anyResult<()> {
    let repos = collect_repos(repo, recursive, verbose)?;
    let statuses = get_statuses(&repos, fetch)?;
    print_status_table(&statuses);
    Ok(())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn get_statuses(repos: &[String], fetch: bool) -> anyResult<Vec<RepoStatus>> {
    let statuses = par_process_repos(repos, |repo| {
        let name = Path::new(repo)
            .file_name()
            .unwrap()
            .to_string_lossy()
            .into_owned();
        get_single_status(repo, &name, fetch)
    });
    Ok(statuses)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn get_single_status(repo_path: &str, name: &str, fetch: bool) -> anyResult<RepoStatus> {
    let porcelain = call_git(repo_path, &["status", "--porcelain"])?;
    let clean = porcelain.is_empty();
    if fetch {
        call_git(repo_path, &["fetch"]).ok();
    }
    let upstream = match call_git(
        repo_path,
        &[
            "rev-parse",
            "--abbrev-ref",
            "--symbolic-full-name",
            "@{upstream}",
        ],
    ) {
        Ok(u) if !u.is_empty() => u,
        _ => {
            return Ok(RepoStatus {
                name: String::new(),
                clean,
                upstream: "—".to_string(),
                ahead: 0,
                behind: 0,
                error: None,
            });
        }
    };
    let rev_out = call_git(
        repo_path,
        &["rev-list", "--count", "--left-right", "@{upstream}...HEAD"],
    )
    .unwrap_or_default();
    let parts: Vec<&str> = rev_out.split_whitespace().collect();
    let (behind, ahead) = if parts.len() == 2 {
        (parts[0].parse().unwrap_or(0), parts[1].parse().unwrap_or(0))
    } else {
        (0, 0)
    };
    Ok(RepoStatus {
        name: name.to_string(),
        clean,
        upstream,
        ahead,
        behind,
        error: None,
    })
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn print_status_table(statuses: &[RepoStatus]) {
    let headers = vec!["Repo", "Clean", "Upstream", "Ahead", "Behind"];
    let align_left = [true, true, true, false, false];
    let rows: Vec<Vec<String>> = statuses
        .iter()
        .map(|s| {
            if let Some(ref err) = s.error {
                return vec![
                    s.name.clone(),
                    format!("error: {}", err),
                    String::new(),
                    String::new(),
                    String::new(),
                ];
            }
            let clean = if s.clean {
                "clean".green().to_string()
            } else {
                "unclean".red().to_string()
            };
            let upstream = if s.upstream == "—" {
                "—".dimmed().to_string()
            } else {
                s.upstream.clone()
            };
            let ahead = if s.ahead > 0 {
                s.ahead.to_string().yellow().to_string()
            } else {
                s.ahead.to_string()
            };
            let behind = if s.behind > 0 {
                s.behind.to_string().yellow().to_string()
            } else {
                s.behind.to_string()
            };
            vec![s.name.clone(), clean, upstream, ahead, behind]
        })
        .collect();

    let mut widths = vec![0; headers.len()];
    for (i, h) in headers.iter().enumerate() {
        widths[i] = h.len();
    }
    for row in &rows {
        for (i, cell) in row.iter().enumerate() {
            widths[i] = widths[i].max(cell.len());
        }
    }
    print_table(&headers, &widths, &align_left, &rows);
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Sync
////////////////////////////////////////////////////////////////////////////////////////////////////

pub struct SyncResult {
    pub name: String,
    pub success: bool,
    pub message: String,
    pub error: Option<String>,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn sync_repos(
    repos: &[String],
    push: bool,
    pull: bool,
    dry_run: bool,
) -> anyResult<Vec<SyncResult>> {
    let results = par_process_repos(repos, |repo| sync_single(repo, push, pull, dry_run));
    Ok(results)
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: refactor & segregate push / pull
fn sync_single(repo_path: &str, push: bool, pull: bool, dry_run: bool) -> anyResult<SyncResult> {
    let name = Path::new(repo_path)
        .file_name()
        .unwrap()
        .to_string_lossy()
        .into_owned();
    let porcelain = call_git(repo_path, &["status", "--porcelain"])?;
    if !porcelain.is_empty() {
        return Ok(SyncResult {
            name,
            success: false,
            message: "Skipped (uncommitted changes)".to_string(),
            error: None,
        });
    }
    let (ahead, behind) = get_ahead_behind(repo_path).unwrap_or((0, 0));
    if pull {
        if dry_run {
            let msg = if behind > 0 {
                format!("[DRY RUN] Would pull {} commits", behind)
            } else {
                "[DRY RUN] Already up to date".to_string()
            };
            return Ok(SyncResult {
                name,
                success: true,
                message: msg,
                error: None,
            });
        }
        let output = Command::new("git")
            .args(["-C", repo_path, "pull"])
            .output()?;
        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);
        if output.status.success() {
            let msg = if stdout.contains("Already up to date") {
                "Already up to date".to_string()
            } else if behind > 0 {
                format!("Pulled {} commits", behind)
            } else {
                let first_line = stdout.lines().next().unwrap_or("");
                if first_line.is_empty() {
                    "Pulled changes".to_string()
                } else {
                    first_line.to_string()
                }
            };
            Ok(SyncResult {
                name,
                success: true,
                message: msg,
                error: None,
            })
        } else {
            let err_msg = if stderr.contains("divergent") {
                "Divergent branches".to_string()
            } else if stderr.contains("no tracking information") {
                "No upstream".to_string()
            } else {
                stderr.lines().next().unwrap_or("Pull failed").to_string()
            };
            Ok(SyncResult {
                name,
                success: false,
                message: err_msg,
                error: None,
            })
        }
    } else {
        // push
        if dry_run {
            let msg = if ahead > 0 {
                format!("[DRY RUN] Would push {} commits", ahead)
            } else {
                "[DRY RUN] Everything up-to-date".to_string()
            };
            return Ok(SyncResult {
                name,
                success: true,
                message: msg,
                error: None,
            });
        }
        let output = Command::new("git")
            .args(["-C", repo_path, "push"])
            .output()?;
        let stdout = String::from_utf8_lossy(&output.stdout);
        let stderr = String::from_utf8_lossy(&output.stderr);
        if output.status.success() {
            let msg = if stdout.contains("Everything up-to-date") {
                "Everything up-to-date".to_string()
            } else if ahead > 0 {
                format!("Pushed {} commits", ahead)
            } else {
                let first_line = stdout.lines().next().unwrap_or("");
                if first_line.is_empty() {
                    "Pushed changes".to_string()
                } else {
                    first_line.to_string()
                }
            };
            Ok(SyncResult {
                name,
                success: true,
                message: msg,
                error: None,
            })
        } else {
            let err_msg = if stderr.contains("divergent") {
                "Divergent branches (pull first)".to_string()
            } else if stderr.contains("no upstream") {
                "No upstream (set with --set-upstream)".to_string()
            } else {
                stderr.lines().next().unwrap_or("Push failed").to_string()
            };
            Ok(SyncResult {
                name,
                success: false,
                message: err_msg,
                error: None,
            })
        }
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn get_ahead_behind(repo_path: &str) -> anyResult<(usize, usize)> {
    let upstream = call_git(
        repo_path,
        &[
            "rev-parse",
            "--abbrev-ref",
            "--symbolic-full-name",
            "@{upstream}",
        ],
    )?;
    if upstream.is_empty() {
        return Ok((0, 0));
    }
    let rev = call_git(
        repo_path,
        &["rev-list", "--count", "--left-right", "@{upstream}...HEAD"],
    )?;
    let parts: Vec<&str> = rev.split_whitespace().collect();
    if parts.len() == 2 {
        Ok((parts[0].parse().unwrap_or(0), parts[1].parse().unwrap_or(0)))
    } else {
        Ok((0, 0))
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn print_sync_table(results: &[SyncResult], action: &str) {
    let headers = vec!["Repo", "Action", "Result", "Message"];
    let align_left = [true, true, true, true];
    let rows: Vec<Vec<String>> = results
        .iter()
        .map(|r| {
            let result_str = if r.success {
                "success".green().to_string()
            } else {
                "failed".red().to_string()
            };
            let mut msg = r.message.clone();
            if msg.contains("uncommitted") || msg.contains("divergent") {
                msg = msg.yellow().to_string();
            } else if msg.contains("Pushed") || msg.contains("Pulled") {
                msg = msg.green().to_string();
            }
            vec![r.name.clone(), action.to_string(), result_str, msg]
        })
        .collect();
    let mut widths = vec![0; headers.len()];
    for (i, h) in headers.iter().enumerate() {
        widths[i] = h.len();
    }
    for row in &rows {
        for (i, cell) in row.iter().enumerate() {
            widths[i] = widths[i].max(cell.len());
        }
    }
    print_table(&headers, &widths, &align_left, &rows);
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn print_table(headers: &[&str], widths: &[usize], align_left: &[bool], rows: &[Vec<String>]) {
    let header_str: Vec<String> = headers
        .iter()
        .enumerate()
        .map(|(i, h)| {
            let aligned = if align_left[i] {
                format!("{:<width$}", h, width = widths[i])
            } else {
                format!("{:>width$}", h, width = widths[i])
            };
            aligned.bold().to_string()
        })
        .collect();
    println!("| {} |", header_str.join(" | "));

    let separator: Vec<String> = widths.iter().map(|w| "-".repeat(*w + 2)).collect();
    println!("|{}|", separator.join("|"));

    for row in rows {
        let formatted: Vec<String> = row
            .iter()
            .enumerate()
            .map(|(i, cell)| {
                if align_left[i] {
                    format!("{:<width$}", cell, width = widths[i])
                } else {
                    format!("{:>width$}", cell, width = widths[i])
                }
            })
            .collect();
        println!("| {} |", formatted.join(" | "));
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

pub fn print_markdown_table<T>(
    headers: &[&str],
    widths: &[usize],
    align_left: &[bool],
    rows: &[T],
    row_fn: fn(&T) -> Vec<String>,
) {
    let header_str: Vec<String> = headers
        .iter()
        .enumerate()
        .map(|(i, h)| {
            let aligned = if align_left[i] {
                format!("{:<width$}", h, width = widths[i])
            } else {
                format!("{:>width$}", h, width = widths[i])
            };
            aligned.bold().to_string()
        })
        .collect();

    println!("| {} |", header_str.join(" | "));
    let separator: Vec<String> = widths.iter().map(|w| "-".repeat(*w + 2)).collect();
    println!("|{}|", separator.join("|"));
    for row in rows {
        let cells = row_fn(row);
        let formatted: Vec<String> = cells
            .iter()
            .enumerate()
            .map(|(i, cell)| {
                if align_left[i] {
                    format!("{:<width$}", cell, width = widths[i])
                } else {
                    format!("{:>width$}", cell, width = widths[i])
                }
            })
            .collect();
        println!("| {} |", formatted.join(" | "));
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn truncate(s: &str, max_chars: usize) -> String {
    if s.chars().count() > max_chars {
        s.chars().take(max_chars).collect()
    } else {
        s.to_string()
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////

fn trim_to_period_or_newline(s: &str) -> String {
    if let Some(idx) = s.find(|c| c == '.' || c == '\n') {
        s[..=idx].trim().to_string()
    } else {
        s.trim().to_string()
    }
}

////////////////////////////////////////////////////////////////////////////////////////////////////
