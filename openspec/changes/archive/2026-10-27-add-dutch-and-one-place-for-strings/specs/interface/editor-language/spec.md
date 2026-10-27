## MODIFIED Requirements

### Requirement: Version control words never reach a content creator

Text a content creator can read SHALL NOT use the words the design hand-off
bans: commit, branch, push, pull, merge, repository, deploy, build log, HEAD,
SHA, and conflict. The plain-language replacements SHALL be used instead.
The rule SHALL hold wherever the strings live, and in every language the
interface has.

#### Scenario: A banned word in a screen

- **WHEN** a string a content creator can read contains a banned word
- **THEN** the frontend check fails, naming the file and the word

#### Scenario: A banned word in the catalogue

- **WHEN** a string in the catalogue contains a banned word, in any language
- **THEN** the frontend check fails, naming the key and the word

#### Scenario: The replacements are used

- **WHEN** the interface refers to saved work, to changes, to publishing, to a
  version, or to going back to one
- **THEN** it uses those words rather than their version control equivalents

#### Scenario: The word is part of a longer, ordinary word

- **WHEN** a string contains a word that merely spells a banned one inside it
- **THEN** the check does not fail
