# gitwrapped

Spotify Wrapped, but for a developer's git history.

Point it at any local git repo and it generates a shareable SVG card showing your busiest coding hour, busiest day of the week, longest commit streak, and your most-changed file — plus a short, witty AI-generated "developer personality" blurb based on your stats.

## What it does

- Reads a repo's commit history (`git log`) and parses it into structured data
- Aggregates stats: commits by hour, commits by weekday, longest daily commit streak, most-changed file by line count
- Renders everything as a polished, dark-themed SVG card with a bar chart histogram
- Sends your aggregated stats (not your code or commit messages) to Google's Gemini API to generate a short, playful personality blurb
- Automatically sizes the card to fit however long that blurb ends up being

## Requirements

- [Go](https://go.dev/dl/) 1.24 or later
- Git installed and available on your `PATH`
- A free [Gemini API key](https://ai.google.dev/gemini-api/docs) (1,500 requests/day free, no card required)

## Setup

1. Clone this repo:
   ```bash
   git clone https://github.com/kevinnethan2008/gitwrapped.git
   cd gitwrapped
   ```

2. Set your Gemini API key as an environment variable:
   ```bash
   export GEMINI_API_KEY="your-key-here"
   ```
   Add this line to your `~/.bashrc` or `~/.zshrc` if you want it to persist across terminal sessions.

## Usage

Run it against any repo by passing the path as an argument:

```bash
go run . /path/to/some/repo
```

Or, to analyze the repo you're currently standing in:

```bash
go run .
```

This produces a file called `wrapped.svg` in the current directory. Open it in a browser or image viewer to see your card.

## Example

```bash
go run . ~/projects/my-old-side-project
```

```
Wrote wrapped.svg
```

Open `wrapped.svg` to see your stats.

## Security note

Only aggregated numbers (counts, hours, a filename) are sent to Gemini — never your raw commit messages or source code.

## Known limitations

- Analyzes only the currently checked-out branch's history from `HEAD`
- The word-wrap on the AI blurb is estimated by character count, not real font-measured pixel width, so wrapping is occasionally a little uneven
