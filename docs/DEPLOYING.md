# Deploying ChatCMS

ChatCMS is deployed with Nix. The flake carries a NixOS module; a machine that
imports it and sets a handful of options runs the whole thing, migrations
included. There is no deployment step that is not in this file.

## The smallest configuration that works

```nix
{
  imports = [ chatcms.nixosModules.default ];

  services.chatcms = {
    enable = true;
    baseUrl = "https://cms.example.org";
    secretKeyFile = "/run/secrets/chatcms-master-key";
    database.createLocally = true;
  };
}
```

That gives a PostgreSQL on the same machine, a `chatcms` user, state under
`/var/lib/chatcms`, and the server on `127.0.0.1:8080` behind whatever you put
in front of it.

`nix/vm-test.nix` boots exactly this, so the configuration above is checked
rather than remembered.

## The master key

Every project's deploy key is encrypted with one master key, so a copy of the
database is worth nothing without it.

- It is 32 bytes of randomness: `head -c 32 /dev/urandom > master.key`.
- `secretKeyFile` names a path on the machine. **Never a path in the Nix
  store**, and never a value in your configuration: anything written in Nix
  ends up in the store, and the store is world readable. Use sops-nix,
  agenix, or a file you put there yourself.
- **Keep it owned by root with mode 0600.** The service does not read it
  directly: systemd reads it as root and hands the service a copy only that
  service can see. You do not have to make it readable by the `chatcms` user,
  and you should not.
- The VM test asserts that the key stays root-owned, that the service still
  starts, and that no copy of its contents is in the store.
- **Losing it loses every deploy key.** The sites are untouched, because they
  are in git, but every project needs a new key added to its git host. Back it
  up somewhere that is not the database's backup.

## What to back up

| What | Why | How |
| --- | --- | --- |
| The master key | Without it every deploy key is unreadable | Wherever you keep secrets, separately from the database |
| PostgreSQL | People, projects, roles, publishes, builds, and the picture on each account | `pg_dump`, as for any database |
| `stateDir` | Working copies and thumbnails | Not strictly needed: working copies are clones and thumbnails are a cache. Backing it up only saves a re-clone. |

Content is never only here. Every page, picture and version of a site is in
the customer's own git remote, which is the point.

A person's picture is the exception, and is deliberately in the database
rather than in any site: it belongs to the account, it crosses
organisations, and a customer's published site must never carry a picture of
somebody from another tenant. It is the one thing a restore brings back that
git cannot.

## Upgrading

Migrations run at start-up, from the binary, before the server listens. There
is no separate step and no migration tool to install.

A migration that fails stops the server rather than starting with a database
it does not understand:

```
level=ERROR msg="the server stopped" error="migrating the database failed: ..."
```

What to do:

1. Nothing has started, so nothing is half-migrated. The old version is still
   able to run against the old schema.
2. Roll back to the previous generation: `nixos-rebuild switch --rollback`.
3. Read the error. It names the migration.
4. Fix it forward in a new release rather than editing an applied migration:
   goose records what it has run, and editing history makes the two disagree.

A rollback to a generation whose schema is older than the database needs the
database rolled back too. Restore the dump you took before upgrading. This is
why the dump is taken before, not after.

## Builds need to be isolated

A site's build runs inside bubblewrap. The module puts it and `prlimit` on the
service's path. If neither is available, builds are refused rather than run
unprotected, and the log says so at start-up:

```
level=WARN msg="builds are unavailable because they cannot be isolated on this machine"
```

A container without user namespaces will say this. `RestrictNamespaces` is
deliberately off in the service for the same reason.

## Sending mail

Invitations and password resets are sent through a mail server you name. With
none named, ChatCMS still works: the link is written to the service log
instead, which suits an installation run by the person who owns it and
nothing larger. Anything with customers on it needs a real one.

```nix
services.chatcms.mail = {
  host = "smtp.example.org";
  port = 587;
  from = "chatcms@example.org";
  username = "chatcms";
  passwordFile = "/var/lib/secrets/chatcms-mail";   # never in the Nix store
};
```

`from` defaults to `chatcms@` the host of `baseUrl`, which is what a
receiving server is most likely to accept. Leave `startTLS` on for anything
reachable over a network; it is only turned off for a development mail
catcher with no certificate.

Sending happens on the job queue, so a mail server that is slow or refusing
delays an invitation rather than losing it, and a failure shows up as a
failed job rather than as a message nobody got.

Check it after deploying by inviting yourself. If nothing arrives, the
service log says whether the message was queued, sent, or refused, and by
whom.

## Appointing the first administrator

Creating an organisation needs somebody who administers the installation, and
a new one has nobody who could appoint anybody. Name an address:

```nix
services.chatcms.firstAdministrator = "you@example.org";
```

The account must already exist. That account is made an administrator every
time the service starts, which is harmless when it already is, and the log
says whether anything changed. Until the account exists, the service starts
and says it could not appoint anybody; the next start finds them.

## Settings worth knowing

Everything is a module option; the full list is in `nix/nixos-module.nix`.
The ones that matter most often:

| Option | What it decides |
| --- | --- |
| `baseUrl` | What invitation links say. Wrong here means invitations nobody can use. |
| `firstAdministrator` | Who may create organisations on a new installation |
| `mail.*` | Where invitations and resets are sent |
| `buildMemory`, `buildTimeout` | What one site's build may use |
| `checkMemory`, `checkTimeout` | What a site's own check may use |
| `previewIdle`, `previewsAtOnce` | How long an unused preview is kept, and how many run |
| `workflowTools` | What a site's check step may run. Fixed and pinned. |

## Behind a reverse proxy

The server speaks plain HTTP and sets secure cookies when `baseUrl` begins
with `https://`. Set it to the address people actually use, terminate TLS in
front, and forward to `listenAddress:port`.

The preview is served under the app's own address, so nothing extra has to be
proxied for it to work.
