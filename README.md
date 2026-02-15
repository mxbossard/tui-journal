# tui-journal

## Purpose
Taking notes everytime, everywhere with multiple formats and in particular with a terminal.
Encrypt notes to be able to push them on a public git repo.
Bullet journal style: manage todo lists, rendez-vous, mindmapping? , ... 
Organize notes by consolidating  topics afterwards, linking subjects, seeking infos through them.

## Security considerations
- https://keepass.info/help/base/security.html#secencrypt
- https://dzone.com/articles/implementing-testing-cryptographic-primitives-go

## Features

## Specs

## Ideas
- Using IPFS ?
  - https://docs.ipfs.tech/concepts/what-is-ipfs/#defining-ipfs 
  - https://github.com/ipfs/kubo

- Numeric bullet journal
  - journaling
  - todolists
  - calendar / agenda ?
  - objectives (life, yearly, monthly, weekly, daily)
  - mood trackers ?
  - habit trackers ?

- Intelligent markdown
  - https://www.markdownlang.com/fr/cheatsheet/
  - auto bullet (enter add another bullet, tab shift sub bullet)
  - shortcut to build todolist
  - help to build dates ?
  - help to find existing topics

- Todolist features
  - Standard not done / done
  - deadline date
  - strike off
  - gamification ?
  - stats ?
  - split tasks into a new todolist ?
  - merge todolists ?
  - Kind
    - daily todolists
    - topic linked todolists ?
  - not done daily todolist jobs duplicated day by day into new todolist until finished ?

- Social interractions
  - Does consigning social interraction need a dedicated feature or topic is sufficient ?
  - carnet d'address
  - template for topics some topics ?

- Time series / Numeric tracker ?
  - Weight
  - Height
  - Rain
  - ...

- Mind Mapping
  - Need to consigne relations between topics
  - 2 topics are related if both mentionned in same paragraph ?
  - How to describe a Mind Map ?
    - &Foo - &Bar - &Baz
  - integration de Mermaid ?


- References patterns
  - topic ref (tag)
    - #topic is not adapted because # is widely used in markdown
    - %topic ?
    - <topic ?
    - !topic ?
    - ?topic ?
    - §topic ?
    - &topic ?
  - date ref
    - @DATE_FORMAT could refer to a date or a deadline
    - 
  - name ref
    - does names need a dedicated ref, or topic is enough ?
    - @TEXT pour mentionner une personne ?

### Encrypted File Layer


### Terminal User Interface
#### Journaling Feature
Objectif: like a numeric bullet journal
- A textarea as main window for Journaling in markdown format
  - As readonly title the current date like: # Vendredi 24/11/2025
  - An help at bottom
  - Some shortcuts to help create todolists, rdv, ...
- A side window to display the formatted markdown
- A side window to display opened todolists

#### Topic consolidation feature
Objectif: organizing my notes
- Prompt one or more topics
- Display ReadOnly journal entries mentionning topics
- Display Topics files if it exists
- Allow to duplicate, rewrite, reorganize data to consolidate Topics

#### Todo list management feature ?
- Which todolist to track ?
- Duplicate a todolist (ex: voyage todolist)
- 

#### Agenda feature ?

### Web User Interface
Running a local web server in option ?

### Use cases
- working offline
  - like git distributed repo: journaling must work offline on each device
  - journaling must be merged automatically without conflicts
  - consolidated topic could be merged automatically with conflict management like git
- working with "patch"
  - offline work could produce a patch file, emailable, to be synced on another repo.
- 

