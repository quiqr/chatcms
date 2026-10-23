// Starts everything an acceptance test needs and stops it again: a PostgreSQL
// of its own, a bare remote on disk, a seeded database and the server. It
// takes nothing from the machine it runs on beyond the programs Nix pins, so
// the suite says the same thing in the gate as on a developer's laptop.
import { execFileSync, spawn } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync, existsSync, mkdirSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { randomBytes } from "node:crypto";

const sh = (command, args, options = {}) =>
  execFileSync(command, args, { encoding: "utf8", ...options });

export class Harness {
  constructor() {
    this.root = mkdtempSync(join(tmpdir(), "chatcms-e2e-"));
    this.pgdata = join(this.root, "pgdata");
    this.pgsock = join(this.root, "sock");
    this.state = join(this.root, "state");
    this.server = null;
  }

  get url() {
    return `http://127.0.0.1:${this.port}`;
  }

  // remoteOf is the address of a project's own remote on disk, which is what
  // an implementer would paste into the setup screen.
  remoteOf(project) {
    return join(this.state, "dev-remotes", `${project}.git`);
  }

  async start() {
    mkdirSync(this.pgsock, { recursive: true });
    mkdirSync(this.state, { recursive: true });

    sh("initdb", ["--username=chatcms", "--auth=trust", "--encoding=UTF8", "--no-locale", "-D", this.pgdata]);
    sh("pg_ctl", ["-D", this.pgdata, "-o", `-k ${this.pgsock} -h ''`, "-l", join(this.root, "pg.log"), "start"]);
    sh("createdb", ["-h", this.pgsock, "-U", "chatcms", "chatcms_e2e"]);

    this.database = `postgres:///chatcms_e2e?host=${this.pgsock}&user=chatcms`;

    // Written as text, not as raw bytes: the server trims whitespace from the
    // key file, so raw bytes that happen to start or end with a space would
    // leave it a byte short, which is a failure one run in twenty.
    const key = join(this.state, "master.key");
    writeFileSync(key, randomBytes(32).toString("hex"), { mode: 0o600 });

    // The seed makes an organisation, people, two projects and one site with
    // a remote of its own, which is what the acceptance tests act on.
    sh(process.env.CHATCMS_SEED ?? "seed", [this.database], {
      env: { ...process.env, CHATCMS_STATE_DIR: this.state },
    });

    this.port = 18080 + Math.floor(Math.random() * 2000);

    this.server = spawn(process.env.CHATCMS_BIN ?? "chatcmsd", [], {
      env: {
        ...process.env,
        CHATCMS_DATABASE_URL: this.database,
        CHATCMS_SECRET_KEY_FILE: key,
        CHATCMS_STATE_DIR: this.state,
        CHATCMS_LISTEN: `127.0.0.1:${this.port}`,
        CHATCMS_BASE_URL: this.url,
      },
      stdio: ["ignore", "pipe", "pipe"],
    });

    this.log = "";
    this.server.stdout.on("data", (chunk) => (this.log += chunk));
    this.server.stderr.on("data", (chunk) => (this.log += chunk));

    await this.waitUntilAnswering();
  }

  // Waits by asking, with a deadline, rather than by sleeping for a guess.
  async waitUntilAnswering() {
    const deadline = Date.now() + 60_000;
    while (Date.now() < deadline) {
      if (this.server.exitCode !== null) {
        throw new Error(`the server stopped before it listened:\n${this.log}`);
      }
      try {
        const answer = await fetch(`${this.url}/api/health`);
        if (answer.ok) return;
      } catch {
        // Not up yet.
      }
      await new Promise((wake) => setTimeout(wake, 100));
    }
    throw new Error(`the server did not answer in time:\n${this.log}`);
  }

  // onTheRemote asks the customer's own remote what it holds, which is how a
  // test checks that publishing actually published.
  onTheRemote(project, ...args) {
    const remote = join(this.state, "dev-remotes", `${project}.git`);
    if (!existsSync(remote)) throw new Error(`no remote at ${remote}`);
    return sh(process.env.CHATCMS_GIT_BIN ?? "git", ["-C", remote, ...args]);
  }

  // projectId is how the seed names the site on disk, which the remote path
  // and the working copy are both named after.
  projectId() {
    const ids = sh("ls", [join(this.state, "sites")]).trim().split("\n");
    return ids[0];
  }

  // pushFromSomewhereElse is somebody editing the site outside ChatCMS, which
  // is what acceptance test 6 is about.
  pushFromSomewhereElse(project, path, content, message) {
    const working = mkdtempSync(join(this.root, "elsewhere-"));
    const remote = join(this.state, "dev-remotes", `${project}.git`);
    const git = process.env.CHATCMS_GIT_BIN ?? "git";

    sh(git, ["clone", remote, working]);
    const where = join(working, path);
    mkdirSync(join(where, ".."), { recursive: true });
    writeFileSync(where, content);
    sh(git, ["-C", working, "add", "--all"]);
    sh(git, ["-C", working, "-c", "user.name=Joris", "-c", "user.email=joris@example.org",
      "commit", "--message=" + message]);
    sh(git, ["-C", working, "push", "origin", "HEAD:refs/heads/main"]);
  }

  async stop() {
    if (this.server && this.server.exitCode === null) {
      this.server.kill("SIGTERM");
      await new Promise((done) => this.server.on("exit", done));
    }
    try {
      sh("pg_ctl", ["-D", this.pgdata, "stop", "-m", "immediate"]);
    } catch {
      // Already gone.
    }
    rmSync(this.root, { recursive: true, force: true });
  }
}

export const owner = { email: "owner@example.org", password: "a long enough passphrase" };
