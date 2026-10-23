# The acceptance tests from docs/BRIEFING-M01.md section 6, run against the
# binary this flake builds, with a PostgreSQL and a git remote of their own.
# Nothing is taken from the machine: every program comes from here, so the
# suite says the same thing in the gate as on a laptop.
{ self, pkgs, system }:

let
  chatcmsd = self.packages.${system}.chatcmsd;

  seed = pkgs.buildGoModule {
    pname = "chatcms-seed";
    inherit (chatcmsd) version src vendorHash;
    subPackages = [ "cmd/seed" ];
  };

  # Without this, a developer's node_modules and test-results go into the
  # store and every local run changes the derivation.
  specs = pkgs.lib.cleanSourceWith {
    src = ../test/e2e;
    filter = path: _type:
      let name = baseNameOf path;
      in name != "node_modules" && name != "test-results" && name != "shots";
  };

  suite = pkgs.buildNpmPackage {
    pname = "chatcms-e2e-deps";
    version = "0.1.0";
    src = specs;
    npmDepsHash = "sha256-RTd1JhjOPx2b40yj8qf5Xt3hVoFN7AuXR814RiGWVb0=";
    dontNpmBuild = true;
    installPhase = ''
      mkdir -p $out
      cp -r node_modules package.json $out/
    '';
  };
in
pkgs.runCommand "chatcms-acceptance-tests"
{
  nativeBuildInputs = [
    pkgs.nodejs_22
    pkgs.postgresql_16
    pkgs.git
    pkgs.hugo
    pkgs.bubblewrap
    pkgs.util-linux
    pkgs.coreutils
  ];

  src = specs;
} ''
  set -euo pipefail

  export HOME=$TMPDIR
  export PLAYWRIGHT_BROWSERS_PATH=${pkgs.playwright-driver.browsers}
  export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
  # Without fonts and a fontconfig to find them by, Chromium dies the moment
  # it has to pick one: "SkFontMgr_FontConfigInterface: Not implemented".
  # The symptom is not a crash message but a page that renders and then
  # ignores every keystroke, which is a bad day to debug.
  export FONTCONFIG_FILE=${pkgs.makeFontsConf { fontDirectories = [ pkgs.dejavu_fonts pkgs.liberation_ttf ]; }}
  export CHATCMS_BIN=${chatcmsd}/bin/chatcmsd
  export CHATCMS_SEED=${seed}/bin/seed
  export CHATCMS_GIT_BIN=${pkgs.git}/bin/git
  export CHATCMS_HUGO_BIN=${pkgs.hugo}/bin/hugo
  export CI=1

  cp -r $src/. .
  chmod -R u+w .
  cp -r ${suite}/node_modules .

  # The two acceptance projects only: the screen capture is a tool for
  # comparing against the drawings, not a check.
  ${pkgs.nodejs_22}/bin/npx playwright test --project=desk --project=phone

  touch $out
''
