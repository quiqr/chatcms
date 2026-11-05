# Briefing: SES and one DNS record for chatcms.eu

For a session in another repository. It lives here because this is the
project that needs the outbox, and because a briefing on somebody's desktop
is a briefing that does not survive moving machine.

A session in `~/gh.snel-city/wasnel-awsaccount-104144963194-main`, doing two
things for ChatCMS: giving it an outbox, and adding the one record its
application host needs. The outbox is modelled on
`060_ses_studio_mint_shop`, which did the same job for `studio-mint.shop`
and is the pattern to copy rather than invent.

Both halves are small, and the second is five lines. Do the DNS record first:
it is quick, it is in a domain that already exists, and it gets you a plan and
an apply through `stackctl` before the SES work asks anything harder of you.

## What this is for

ChatCMS will run on durer under `chatcms.eu` and must send invitations and
password resets. Its NixOS module takes `mail.host`, `mail.port`,
`mail.username`, `mail.passwordFile`, `mail.from`, `mail.fromName` and
`mail.startTLS`. So this session produces three things: a verified,
DKIM-signed identity for `chatcms.eu`, an SMTP user that may send only for
it, and that user's password. Nothing is deployed and nothing outside this
repository changes.

## The ground

Two IaC trees live side by side and share nothing:

- `stack/` is Terraform, being handed over one domain at a time.
- `nivis/` is nivis 0.10, Nix, state in
  `s3://nivis-state-104144963194-production`.

**This work is entirely in `nivis/`.** `stack/r53_zones/r53_zones/studio-mint.shop/ses.tf`
still exists and is the Terraform ancestor of what you are copying. Read it
if you like, but do not extend it, and do not copy its `aws_ses_domain_mail_from`
block: the nivis version deliberately left the custom mail-from domain out,
and with it the SPF and MX records it needs. Bounces go to the amazonses
default. Keep that simplification.

Account 104144963194, region `eu-central-1`, profile passed in the
environment because the repository names none.

```bash
nix develop ./nivis          # pinned nivis and awscli2
AWS_PROFILE=wasnel ./nivis/stackctl production <domain> plan
AWS_PROFILE=wasnel ./nivis/stackctl production <domain> apply
```

`chatcms.eu` is already in code as a nivis zone:
`nivis/stack/010_route53/zones/chatcms.eu.nix`, holding the zone itself, an
apex `A` and a `www` `A` at `178.104.252.76` (durer), and two records for an
ecosymbia staging site. Nothing needs importing.

## Part one: the application host

ChatCMS runs on durer, with the application on `app.chatcms.eu` and the apex
serving a static front page. The apex and `www` already point at durer; the
application subdomain does not exist yet.

**The name already resolves.** `dig +short app.chatcms.eu` answers
`178.104.252.76` today, it is not a wildcard, and no record for it is in
`zones/chatcms.eu.nix`. Somebody made it in the console, so this is drift:
the thing exists, and the code does not know about it.

That changes the task from adding a record to adopting one. Decide which:
either `allow_overwrite = true` on the record, or an entry in
`nivis/import/010_route53.imports.json` the way `060_ses_studio_mint_shop`
adopted its own. Creating it outright will fail, because Route53 refuses to
create a record that is already there.

Add it to `nivis/stack/010_route53/zones/chatcms.eu.nix`, in the shape the
two records above it already use:

```nix
(record {
  name = "chatcms_eu_app";
  recordName = "app.chatcms.eu";
  type = "A";
  ttl = 3600;
  records = [ "178.104.252.76" ];
})
```

Then `plan` on `010_route53` must touch nothing but that one record, and
leave the apex, `www` and the two ecosymbia records exactly as they are.

## The sandbox, which is behind us

On 30 October the account was in the SES sandbox:
`ProductionAccessEnabled: false`, 200 mails a day to verified recipients
only. The studio-mint change requested production access and it was granted;
Pim confirmed that on 5 November.

Worth one command at the start anyway, because the whole session rests on it:

```bash
AWS_PROFILE=wasnel aws sesv2 get-account --region eu-central-1 \
  --query '{production:ProductionAccessEnabled,quota:SendQuota}'
```

## Part two: the outbox

### 1. A new domain, registered in three places

Copy `nivis/stack/060_ses_studio_mint_shop/domain.nix` to
`nivis/stack/070_ses_chatcms_eu/domain.nix`, with `domain = "chatcms.eu"`,
every resource name rewritten from `studio_mint_shop` to `chatcms_eu`, the
IAM user named `chatcms-smtp`, its policy named `send-as-chatcms-eu`, and the
backend key `070_ses_chatcms_eu/state.json`.

Then register it:

- `nivis/flake.nix`, the `domains` attribute set
- `nivis/stackctl`, the `domains` array
- `nivis/README.md`, the tree listing near the top

A domain missing from `stackctl` is refused with the valid choices, which is
a good error but only after you have wondered why.

### 2. The DKIM tokens, which cannot be written in advance

This is the one thing that will catch you out.

`dkimTokens` in `domain.nix` is a literal list of three strings. For
studio-mint that was honest: those resources already existed and were adopted
through `nivis/import/060_ses_studio_mint_shop.imports.json`, so the tokens
were known before the file was written.

`chatcms.eu` has no identity yet, so there are no tokens until AWS makes
them. The domain therefore has to be applied in two passes:

1. Write `domain.nix` with `dkimTokens = [ ]`, so the identity, the DKIM
   resource, the verification `TXT`, and the three IAM resources are created
   but no `_domainkey` records are.
2. Read the tokens back:
   ```bash
   AWS_PROFILE=wasnel aws ses get-identity-dkim-attributes \
     --identities chatcms.eu --region eu-central-1 \
     --query 'DkimAttributes."chatcms.eu".DkimTokens'
   ```
3. Put those three strings in `dkimTokens`, plan, and apply again. The three
   `CNAME` records appear, and verification follows within minutes.

Do not invent an import manifest for this. `import/` is for adopting
resources that already exist, which these do not.

### 3. The credentials

The studio-mint domain already builds them and the copy inherits all three:

- `aws_iam_user` `chatcms-smtp`
- `aws_iam_user_policy` allowing only `ses:SendRawEmail` on the
  `chatcms.eu` identity ARN, with a `StringLike` condition on
  `ses:FromAddress` of `*@chatcms.eu`
- `aws_iam_access_key`, whose `ses_smtp_password_v4` attribute is the SMTP
  password

The SMTP endpoint is `email-smtp.eu-central-1.amazonaws.com` on port 587 with
STARTTLS.

### 4. Say where the records come from

`nivis/README.md` already notes, for `studio-mint.shop`, that four records in
that zone are not declared by `010_route53` because `060_ses_studio_mint_shop`
owns the identity whose values they carry. Add the same note for
`chatcms.eu`, or the next person to read `zones/chatcms.eu.nix` will think
the zone has drifted.

## How this repository works

It is OpenSpec and beans, like the others. A change here gets a proposal, a
spec delta and tasks, and `openspec/specs/` already holds
`nivis-repo-layout`, `nivis-route53` and `nivis-state-backend`, which this
work touches. Read them before proposing: the layout spec in particular says
what a nivis domain owes.

## What this session hands over

Two values:

- the SMTP user name (the access key id)
- the SMTP password (`ses_smtp_password_v4`)

Everything on the mipnix side is already in place and waiting for them, as of
`ca867980`:

- `secrets/chatcms-smtp-password.age` exists, encrypted to pim and durer, and
  currently holds a placeholder. Replace it with
  `printf '%s' '<the password>' | agenix -e chatcms-smtp-password.age`, run
  from `mipnix/secrets`. Note that agenix takes stdin whenever stdin is not a
  terminal, so an `EDITOR=` trick is silently ignored and encrypts nothing:
  a 322 byte output means the payload was empty.
- `modules/HOSTS/durer-server/configuration.nix` carries the `mail` block
  commented out, with the host, port, startTLS and from already filled in.
  Put the user name in it and uncomment.

The sending address is `noreply@chatcms.eu`, matching what studio-mint uses.

## Not in this session

- A custom mail-from domain, SPF and the MX record it needs. The nivis
  studio-mint domain does without, and matching it keeps the two comparable.
- DMARC beyond whatever the zone has. studio-mint stayed at `p=none` until a
  few weeks of reports had come in.
- Bounce and complaint handling through SNS. Worth doing before a customer
  sends much; it does not block the first invitation.

## How to know it worked

- `aws sesv2 get-account` says production access is enabled.
- `aws ses get-identity-verification-attributes --identities chatcms.eu`
  says `Success`.
- `aws ses get-identity-dkim-attributes --identities chatcms.eu` says
  `Success` and lists the same three tokens that are in `domain.nix`.
- `dig +short app.chatcms.eu` answers `178.104.252.76`.
- `AWS_PROFILE=wasnel ./nivis/stackctl production 070_ses_chatcms_eu plan`
  is clean, and so is `010_route53`.
- A mail sent through `email-smtp.eu-central-1.amazonaws.com:587` with the
  new credentials arrives, and its headers show `dkim=pass` and `spf=pass`.
