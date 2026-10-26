## Purpose

Defines what ChatCMS sends, how it is sent, what happens when it cannot be,
and what a message may never contain.

## ADDED Requirements

### Requirement: Messages are sent through a configured mail server

The server SHALL send over SMTP to a configured host, as a configured
address, with credentials and an encrypted connection where the server wants
them. Every setting SHALL be readable from a file as well as from a value, so
that a password stays out of the Nix store.

#### Scenario: A message is sent

- **WHEN** a message is handed over and a mail server is configured
- **THEN** it SHALL be sent to that server

#### Scenario: An encrypted connection is asked for and not offered

- **WHEN** the connection is required to be encrypted and the server does not
  offer it
- **THEN** nothing SHALL be sent
- **AND** the reason SHALL say so

#### Scenario: A password from a file

- **WHEN** the password is given as a file
- **THEN** it SHALL be read from that file

### Requirement: Nothing is sent while somebody waits

Sending SHALL happen on the job queue rather than during the request that
caused it.

#### Scenario: Inviting somebody

- **WHEN** somebody is invited
- **THEN** the screen SHALL answer without waiting for the mail server

#### Scenario: The mail server is refusing

- **WHEN** sending fails
- **THEN** it SHALL be tried again
- **AND** the invitation SHALL NOT be lost

#### Scenario: A message that cannot be read

- **WHEN** a queued message cannot be understood
- **THEN** it SHALL be said out loud and dropped rather than tried forever

### Requirement: With nowhere to send, the link is still recorded

An installation with no mail server configured SHALL keep working, and SHALL
write the link where whoever runs it can read it.

#### Scenario: No mail server configured

- **WHEN** somebody is invited and no mail server is configured
- **THEN** the link SHALL be written to the log
- **AND** the log SHALL say that it was not sent, and why

#### Scenario: Starting with none

- **WHEN** the server starts with no mail server configured
- **THEN** it SHALL start, and SHALL say that links will go to the log

### Requirement: A message reads for the person it is sent to

A message SHALL be written in the language the design's copy rules require,
and SHALL NOT contain a word from the vocabulary editors are never shown.

#### Scenario: An invitation

- **WHEN** an invitation is composed
- **THEN** it SHALL say who invited them and to what, where that is known
- **AND** SHALL carry the link in its plain text as well as its HTML

#### Scenario: Words an editor is never shown

- **WHEN** any message is composed
- **THEN** it SHALL NOT contain commit, branch, repository, deploy, push or
  merge

#### Scenario: Nothing to name

- **WHEN** neither the site nor the inviter is known
- **THEN** the message SHALL still read as a sentence, with no gap where a
  name would be

### Requirement: A message cannot carry a header it was not given

What is written into a message SHALL NOT be able to add headers of its own.

#### Scenario: A line break in an address

- **WHEN** a recipient address contains a line break
- **THEN** sending SHALL be refused

#### Scenario: A line break in a subject or a name

- **WHEN** a subject or a display name contains a line break
- **THEN** the message SHALL carry no header that was not meant

#### Scenario: A link in the HTML part

- **WHEN** a link contains characters that would close an attribute
- **THEN** the HTML SHALL NOT gain a tag from it
