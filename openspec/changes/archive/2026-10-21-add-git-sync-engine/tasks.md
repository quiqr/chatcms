## 1. The runner, and the rule that only it runs git

- [x] 1.1 Add `internal/gitsync` with a `runner` that builds the argument
  list, sets the deterministic settings from `design.md`, and executes the
  configured binary; verify with `TestRunnerUsesTheConfiguredBinary` that a
  binary named by configuration is used and the path is not searched.
- [x] 1.2 Make the runner non-interactive and hook-free: no credential helper,
  no terminal prompt, no system or global config, hooks disabled; verify with
  `TestRunnerIsNonInteractive` by pointing at a repository with a hook that
  would fail, and with `TestRunnerIgnoresGlobalConfig`.
- [x] 1.3 Turn a git failure into an error that names the operation and
  carries git's own message with anything credential-shaped removed; verify
  with `TestErrorScrubsCredentials` for a URL carrying a token.
- [x] 1.4 Add `TestOnlyGitsyncRunsGit`, which walks the Go sources and fails
  when a package other than `internal/gitsync` executes git; verify by adding
  a temporary offending file and seeing it fail.

## 2. The working copy

- [x] 2.1 Implement `Clone` into an empty directory, recording the remote;
  verify with `TestCloneBringsContent` against a bare repository on disk.
- [x] 2.2 Implement `Open` with no network access, and make opening a
  non-working-copy fail without changing the directory; verify with
  `TestOpenTouchesNothing` and `TestOpenRejectsAPlainDirectory`.
- [x] 2.3 Make `Clone` refuse a directory that already holds a working copy;
  verify with `TestCloneRefusesToOverwrite`.
- [x] 2.4 Expose content as a read-only `fs.FS` that excludes version
  control's own directory; verify with `TestContentFSReadsFiles` and
  `TestContentFSHidesVersionControl`.
- [x] 2.5 Make the content filesystem refuse a path that is in conflict rather
  than serving a file with markers; verify with
  `TestContentFSRefusesAConflictedPath`.

## 3. Saving

- [x] 3.1 Implement `Save` writing files and producing exactly one version
  with the given message, attributed to the named person with the system as
  the agent; verify with `TestSaveMakesOneVersion` and
  `TestSaveIsAttributedToTheEditor`.
- [x] 3.2 Report a save that changes nothing as no version rather than an
  error; verify with `TestSaveWithNoChangeMakesNoVersion`.
- [x] 3.3 Support deletions in a save; verify with `TestSaveCanDelete`.
- [x] 3.4 Reject a path that escapes the working copy, writing nothing; verify
  with `TestSaveRejectsAnEscapingPath` for `../`, an absolute path and a
  symlink that points outside.
- [x] 3.5 Serialise writes per site with a per-site lock, and confirm two
  sites do not block each other; verify with `TestConcurrentSavesToOneSite`
  and `TestConcurrentSavesToTwoSites`.

## 4. History and status

- [x] 4.1 Implement `History` returning versions newest first with an
  identifier, time, author, message and changed paths; verify with
  `TestHistoryIsNewestFirst` and `TestVersionNamesWhatItChanged`.
- [x] 4.2 Return an empty history on a repository with no versions rather than
  failing; verify with `TestHistoryOnAnEmptyRepository`.
- [x] 4.3 Implement `Status` answering unsaved work, unpublished count and
  whether the remote moved, computed from the repository; verify with
  `TestStatusAfterSave`, `TestStatusCountsUnpublished` and
  `TestStatusSeesTheRemoteMove`.
- [x] 4.4 Confirm status survives having any cached state discarded; verify
  with `TestStatusIsNotRemembered`.

## 5. Exchanging with the remote

- [x] 5.1 Implement `Sync` as fetch then replay of local versions onto the
  remote, reporting what arrived; verify with
  `TestSyncBringsRemoteVersions` and `TestSyncReportsNothingNew`.
- [x] 5.2 Keep the working copy usable when the remote does not answer, and
  report the failure with the remote named and no credential; verify with
  `TestSyncSurvivesAnUnreachableRemote`.
- [x] 5.3 Replay without creating a merge version, with the remote's version
  first; verify with `TestReplayKeepsHistoryLinear`.
- [x] 5.4 Implement `Publish` pushing local versions, and recover from a
  rejected push by syncing and replaying; verify with
  `TestPublishSendsVersions` and `TestRejectedPushLosesNothing`.

## 6. Conflicts

- [x] 6.1 Detect a conflict, read all three versions of each conflicted path
  from the index, and reset the working tree so no marker is on disk; verify
  with `TestConflictCarriesThreeVersions` and
  `TestNoConflictMarkerReachesContent`.
- [x] 6.2 Leave non-overlapping changes alone; verify with
  `TestDifferentFilesDoNotConflict`.
- [x] 6.3 Implement `Resolve` taking content per path and completing the
  exchange; verify with `TestResolveKeepingLocal`,
  `TestResolveKeepingRemote` and `TestResolveWithNewContent`.
- [x] 6.4 Reject a resolution naming a path that is not conflicted, leaving
  the exchange as it was; verify with `TestResolveRejectsAnUnconflictedPath`.
- [x] 6.5 Implement `Abandon` restoring the pre-exchange state with no version
  lost; verify with `TestAbandonRestoresTheWorkingCopy` and
  `TestAbandonLosesNoVersion`.
- [x] 6.6 Remember resolutions within one exchange so the same conflict is not
  asked about twice during a replay; verify with
  `TestReplayDoesNotAskTwiceForTheSameConflict`.
- [x] 6.7 Report an exchange in progress when a working copy is opened after
  an interruption; verify with `TestOpenReportsAnExchangeInProgress` by
  starting an exchange, discarding the handle and opening again.

## 7. Background jobs

- [x] 7.1 Add `internal/jobs` with queueing that returns at once and survives
  a restart; verify with `TestQueueReturnsImmediately` and
  `TestQueuedJobSurvivesARestart`.
- [x] 7.2 Make claiming exclusive, and report an empty queue rather than
  failing; verify with `TestTwoWorkersOneJob` and `TestClaimOnAnEmptyQueue`.
- [x] 7.3 Implement retries with exponential backoff and an attempt limit,
  recording the last reason on final failure; verify with
  `TestFailedJobIsRetriedLater` and `TestRepeatedFailureIsRecorded`.
- [x] 7.4 Scrub credentials from a recorded reason; verify with
  `TestRecordedReasonCarriesNoSecret`.
- [x] 7.5 Deduplicate by kind and subject, with one extra job allowed while
  one is running; verify with `TestQueueingTwiceQueuesOnce` and
  `TestQueueingDuringARunQueuesOneMore`.
- [x] 7.6 Implement a worker that stops claiming on shutdown, lets a running
  job finish within the budget, and returns unfinished work to the queue;
  verify with `TestWorkerDrainsOnShutdown` and
  `TestWorkerReturnsAnOverrunningJob`.
- [x] 7.7 Add the sync job kind, queued on a schedule and on demand; verify
  with `TestSyncJobRunsASync`.

## 8. Wiring and the gate

- [x] 8.1 Start the worker from `internal/app` and drain it on shutdown
  alongside in-flight requests; verify with
  `TestShutdownDrainsTheWorkerToo`.
- [x] 8.2 Add the configuration the engine needs: the sync interval, the
  working copy root under the state directory, and the SSH key path; verify
  with `TestLoadConfigReadsSyncSettings`.
- [x] 8.3 Add a test helper that builds a bare repository and a working copy
  on disk, so every exchange test uses real git; verify every test in groups
  5 and 6 runs against it with no mock.
- [x] 8.4 Make the integration check in `flake.nix` provide the git binary and
  confirm it is the pinned one; verify `nix flake check` passes and the test
  asserting the git version runs.
- [x] 8.5 Confirm `internal/gitsync` is at or above the 80% core-package
  floor; verify the gate no longer reports it as absent.
- [x] 8.6 Add bullets to `CHANGELOG.md` under `## [Unreleased]` written from a
  content creator's perspective, using no git vocabulary.
- [x] 8.7 Note in `docs/ARCHITECTURE.md` that hooks are disabled in customer
  repositories and that workflows are the supported way to run a step; verify
  the sentence is there.
- [x] 8.8 Run `nix flake check` and confirm every check passes.
