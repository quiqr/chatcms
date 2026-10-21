{
  description = "Quiqr ChatCMS: hosted, Nix-powered headless CMS for Markdown static sites";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      version = "0.1.0";

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = f:
        nixpkgs.lib.genAttrs systems (system: f {
          inherit system;
          pkgs = nixpkgs.legacyPackages.${system};
          lib = nixpkgs.lib;
        });

      goSource = nixpkgs.lib.fileset.toSource {
        root = ./.;
        fileset = nixpkgs.lib.fileset.unions [
          ./go.mod
          ./go.sum
          ./cmd
          ./db
          ./internal
          ./test/fixtures
        ];
      };

      webSource = nixpkgs.lib.fileset.toSource {
        root = ./web;
        fileset = nixpkgs.lib.fileset.unions [
          ./web/package.json
          ./web/package-lock.json
          ./web/tsconfig.json
          ./web/vite.config.ts
          ./web/vitest.setup.ts
          ./web/index.html
          ./web/src
          ./web/public
        ];
      };

      vendorHash = "sha256-2K9m6EUifEBZ/iFw1ioE1ef5PqzlvgPHKVs61C3h0ss=";

      mkWeb = { pkgs, ... }: pkgs.buildNpmPackage {
        pname = "chatcms-web";
        inherit version;
        src = webSource;
        npmDepsHash = "sha256-FQsbsYTdcxp2nAKEwU0u+UVVPEakhli2lkqio0cPEg4=";

        nativeBuildInputs = [ pkgs.nodejs_22 ];

        env.ESBUILD_BINARY_PATH = pkgs.lib.getExe pkgs.esbuild;
        env.ROLLUP_SKIP_NODEJS_NATIVE = "1";

        doCheck = true;
        checkPhase = ''
          runHook preCheck
          npm run typecheck
          npm run test
          runHook postCheck
        '';

        installPhase = ''
          runHook preInstall
          mkdir -p $out
          cp -r dist/* $out/
          runHook postInstall
        '';

        meta.description = "Quiqr ChatCMS browser application";
      };

      mkServer = { pkgs, ... }@args: pkgs.buildGoModule {
        pname = "chatcmsd";
        inherit version vendorHash;
        src = goSource;
        subPackages = [ "cmd/chatcmsd" ];
        doCheck = false;
        env.CGO_ENABLED = 0;

        preBuild = ''
          cp -r ${mkWeb args}/. internal/webui/dist/
        '';

        ldflags = [
          "-s"
          "-w"
          "-X=github.com/quiqr/chatcms/internal/version.Version=${version}"
        ];

        meta = {
          description = "Quiqr ChatCMS server";
          mainProgram = "chatcmsd";
        };
      };

      mkGoTests = { pkgs, ... }@args: (mkServer args).overrideAttrs (old: {
        pname = "chatcmsd-tests";
        subPackages = null;
        preBuild = "";
        doCheck = true;
        env = old.env // { CGO_ENABLED = 1; };
        buildPhase = "runHook preBuild; runHook postBuild";
        checkPhase = ''
          runHook preCheck
          go vet ./...
          go test -race -covermode=atomic -coverprofile=coverage.out ./...
          runHook postCheck
        '';
        installPhase = ''
          runHook preInstall
          mkdir -p $out
          cp coverage.out $out/coverage.out
          runHook postInstall
        '';
      });

      mkIntegrationTests = { pkgs, ... }@args: (mkGoTests args).overrideAttrs (old: {
        pname = "chatcmsd-integration-tests";
        nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ pkgs.postgresql_16 ];
        checkPhase = ''
          runHook preCheck
          export HOME=$TMPDIR
          go vet -tags=integration ./...
          go test -tags=integration -covermode=atomic -coverprofile=coverage.out ./...
          ${pkgs.bash}/bin/bash ${./scripts/coverage-gate.sh} coverage.out
          runHook postCheck
        '';
      });

      mkSqlcCheck = { pkgs, ... }: pkgs.runCommand "chatcms-sqlc-diff"
        {
          nativeBuildInputs = [ pkgs.sqlc ];
          src = nixpkgs.lib.fileset.toSource {
            root = ./.;
            fileset = nixpkgs.lib.fileset.unions [
              ./sqlc.yaml
              ./db
              ./internal/store
            ];
          };
        } ''
        cp -r $src/. .
        chmod -R u+w .
        sqlc diff
        touch $out
      '';

      mkDevShell = { pkgs, system, ... }: pkgs.mkShell {
        name = "chatcms";
        packages = with pkgs; [
          go
          gopls
          gotools
          go-tools
          golangci-lint
          delve
          air

          nodejs_22
          typescript-language-server

          postgresql_16
          sqlc
          goose

          hugo
          git
          jujutsu
          openssh

          playwright-driver.browsers
          prefetch-npm-deps
          jq
          yq-go
        ];
        env = {
          PLAYWRIGHT_BROWSERS_PATH = "${pkgs.playwright-driver.browsers}";
          PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD = "1";
          CHATCMS_HUGO_BIN = pkgs.lib.getExe pkgs.hugo;
          CHATCMS_GIT_BIN = pkgs.lib.getExe pkgs.git;
        };
        shellHook = ''
          echo "Quiqr ChatCMS dev shell (${system})"
          echo "  scripts/dev-db.sh up     start a local Postgres in ./.pgdata"
          echo "  nix flake check          build + tests + coverage gate"
          echo "  bash scripts/ship-change.sh <change>  ship one OpenSpec change"
        '';
      };
    in
    {
      packages = forAllSystems (args: rec {
        chatcmsd = mkServer args;
        web = mkWeb args;
        default = chatcmsd;
      });

      devShells = forAllSystems (args: {
        default = mkDevShell args;
      });

      checks = forAllSystems (args: {
        go-build = mkServer args;
        go-test = mkGoTests args;
        integration = mkIntegrationTests args;
        frontend = mkWeb args;
        sqlc = mkSqlcCheck args;
      });

      formatter = forAllSystems ({ pkgs, ... }: pkgs.nixpkgs-fmt);

      nixosModules.default = import ./nix/nixos-module.nix { inherit self; };
      nixosModules.chatcms = self.nixosModules.default;
    };
}
