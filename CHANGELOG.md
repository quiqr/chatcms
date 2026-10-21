# Changelog

All notable changes to Quiqr ChatCMS are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Project scaffold: OpenSpec, beans, Nix flake, ship gate and the milestone 01 backlog.
- The server runs. `chatcmsd` listens, serves a JSON API, reports its health at
  `/api/health`, and stops gracefully without dropping a request in flight.
- The database schema is created and upgraded by the server itself on start-up,
  safely even when two instances start at the same time. A server older than the
  database it is pointed at refuses to start rather than serving against a schema
  it does not understand.
- A deployment is one binary. The browser application ships inside `chatcmsd`, so
  there is no separate directory of assets to copy or keep in step.
- Configuration values can be supplied from files rather than environment
  variables, so passwords and keys stay out of the Nix store.
- Quiqr models are read the way the desktop app reads them: a base file in
  YAML, TOML or JSON, plus includes, per-item include folders and partials,
  assembled in the same order and with the same precedence.
- All twenty-seven field types of the desktop app are understood, and a field
  type this version does not know is kept as a placeholder rather than dropped,
  so a model written for a newer Quiqr still opens.
- An unusable model says which file and which part of it is wrong.
- The pages and collections a content creator browses, with item counts and
  the model's own labels, are derived from the model itself.
