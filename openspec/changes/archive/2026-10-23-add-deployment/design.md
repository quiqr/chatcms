# Design

Epic: `.beans/chatcms-o5xn--deployment-nixos-module-and-release.md`

## The test found the thing a first deployment would have found

The virtual machine test failed on its first run, exactly as somebody's first
deployment would have:

```
"the master key could not be read: open /run/chatcms-secrets/master.key: permission denied"
```

A key owned by root with mode 0600, and a service running as `chatcms`. The
obvious fix is to make the key readable by the service user, which means
widening the permissions on a secret, which is the wrong direction entirely.

systemd already solves this. `LoadCredential=` makes systemd read the file as
root and hand the service a copy only that service can see. The key stays
root-owned, the service stays non-root, and nothing in the middle is widened.

The test now asserts all three: the key is still `root 600`, the service
started anyway, and no copy of its contents is in the store.

## Absence from the store is searched for, not assumed

A path outside the store proves nothing if something copied the contents in.
The test searches the store for the key's own bytes. That is the assertion
worth having, because the failure it guards against is a configuration that
interpolates a secret into Nix without anybody noticing.

## Why the check is Linux only

A virtual machine cannot be booted on a Mac. The check is added only where it
can run; everything else still runs everywhere, so a Mac developer gets the
same gate minus the one check their machine cannot perform.

## What the documentation is for

`docs/DEPLOYING.md` answers the questions somebody has while deploying, in the
order they have them: what is the smallest thing that works, what do I do with
the key, what do I back up, and what do I do when a migration fails. The
migration answer matters most: nothing is half-applied, because the server
refuses to listen, so the recovery is a rollback and a read of the error
rather than a repair of the database.
