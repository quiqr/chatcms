# history/going-back Specification

## Purpose
Defines what going back to an earlier version does, what a person is told
first, and why it is safe.

## Requirements

### Requirement: Going back makes a new version rather than removing any

Going back SHALL make a new version whose content is the earlier version's.
No version SHALL be removed from the history.

#### Scenario: the site holds the earlier content again

- **WHEN** somebody goes back to an earlier version
- **THEN** the site SHALL hold what that version held
- **AND** files the earlier version did not have SHALL no longer be in the site

#### Scenario: the history keeps everything

- **WHEN** somebody goes back
- **THEN** every version that was in the history before SHALL still be in it
- **AND** there SHALL be one more version than before

#### Scenario: going back can itself be undone

- **WHEN** somebody goes back and then goes back to the version they moved
  away from
- **THEN** the site SHALL hold that version's content again

### Requirement: Unpublished work is set aside, not lost

Where somebody goes back past work that has not been published, that work
SHALL stop being in the site but SHALL remain in the history.

#### Scenario: what happens to it

- **WHEN** somebody goes back past unpublished work
- **THEN** that work SHALL no longer be in the site
- **AND** the version holding it SHALL still be in the history

#### Scenario: getting it back

- **WHEN** somebody goes back to the version that held the set-aside work
- **THEN** that work SHALL be in the site again

### Requirement: The confirmation says what will happen

Before going back, a person SHALL be told which version they are going back
to, when it was, how much unpublished work will be set aside, and that nothing
is lost.

#### Scenario: nothing waiting

- **WHEN** nothing is waiting to be published
- **THEN** the confirmation SHALL still say that nothing is lost and that this
  can be undone

#### Scenario: work waiting

- **WHEN** work is waiting to be published
- **THEN** the confirmation SHALL say how much will be set aside

#### Scenario: a version that is not there

- **WHEN** the confirmation is asked for a version the site does not have
- **THEN** the answer SHALL be that there is nothing there

### Requirement: Publishing right away is offered and is on by default

Going back SHALL offer to publish the site straight afterwards, and SHALL do
so unless somebody turns it off.

#### Scenario: publishing right away

- **WHEN** somebody goes back without saying otherwise
- **THEN** the site SHALL be published straight afterwards

#### Scenario: not publishing right away

- **WHEN** somebody turns publishing right away off
- **THEN** the site SHALL go back and SHALL NOT be published
- **AND** the change SHALL be waiting to be published like any other

#### Scenario: the publish fails

- **WHEN** going back works but publishing afterwards fails
- **THEN** the answer SHALL still say that the site went back

### Requirement: Going back to where the site already is is refused

Going back to a version whose content the site already holds SHALL be refused
rather than making a version that changes nothing.

#### Scenario: nothing would change

- **WHEN** somebody goes back to a version whose content the site already
  holds
- **THEN** it SHALL be refused with a reason worth showing
- **AND** no version SHALL be made

### Requirement: What is said about going back carries no technical word

Everything written about going back, including what the new version is called,
SHALL be in the words a content creator uses.

#### Scenario: the version it makes

- **WHEN** going back makes a version
- **THEN** what that version is called SHALL name the version gone back to
- **AND** SHALL carry no git word

### Requirement: Only somebody allowed may go back

Whether somebody may go back SHALL be decided by the policy function, and a
refusal SHALL leave the site unchanged.

#### Scenario: a viewer tries

- **WHEN** somebody who may read but not change the site asks to go back
- **THEN** they SHALL be refused
- **AND** the site SHALL be unchanged
