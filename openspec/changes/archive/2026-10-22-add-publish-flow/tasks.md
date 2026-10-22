# Tasks

## 1. What the git engine had to learn

- [x] 1.1 `Changes`: what differs from the published site, with who and when
- [x] 1.2 `ReadPublished`: a file as the live site has it
- [x] 1.3 `PublishTo`: the site to a named line
- [x] 1.4 `PublishPages`: built pages to a line of their own, through git's
      plumbing, with the working copy untouched
- [x] 1.5 Refuse an unusable line, and built pages over the site's own line
- [x] 1.6 Tests against a real remote for every one of them

## 2. What changed, as pages and fields

- [x] 2.1 Locate a file in the model: page, collection item, image, other
- [x] 2.2 Name a page and a field the way the model names them
- [x] 2.3 Compare front matter field by field, and the page's own text
- [x] 2.4 The word diff, with the read-it-back-both-ways test
- [x] 2.5 Order and group the changes

## 3. Publishing

- [x] 3.1 The publisher that carries out a workflow's publishing step
- [x] 3.2 The service: review, run, read, history
- [x] 3.3 Policy at every entry, through one place
- [x] 3.4 The record: migration, queries, one publish at a time
- [x] 3.5 Mark publishes left running as interrupted on start-up
- [x] 3.6 `sitebuild.Pages`: the built site rather than the folder holding it

## 4. The HTTP layer

- [x] 4.1 Review, publish, read one, list them
- [x] 4.2 The raw output only for those allowed technical detail
- [x] 4.3 Wire it into the server

## 5. The screen

- [x] 5.1 Review: the count, the last publish, the list with badges
- [x] 5.2 The inline word diff, expandable
- [x] 5.3 The note
- [x] 5.4 Done and failed, with the steps
- [x] 5.5 Technical details behind a disclosure, implementer-only
- [x] 5.6 The sticky publish bar on a phone
- [x] 5.7 A test that no git word reaches what an editor reads

## 6. Shipping

- [x] 6.1 An integration test for acceptance test 7, both forms
- [x] 6.2 A test that editing during a publish spoils neither
- [x] 6.3 `docs/ARCHITECTURE.md`
- [x] 6.4 The gate green, coverage held
