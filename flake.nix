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
        });

      goSource = nixpkgs.lib.fileset.toSource {
        root = ./.;
        fileset = nixpkgs.lib.fileset.unions [
          ./go.mod
          ./cmd
          ./internal
        ];
      };

      mkServer = { pkgs, ... }: pkgs.buildGoModule {
        pname = "chatcmsd";
        inherit version;
        src = goSource;
        vendorHash = null;
        subPackages = [ "cmd/chatcmsd" ];
        doCheck = false;
        env.CGO_ENABLED = 0;
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
        doCheck = true;
        env = old.env // { CGO_ENABLED = 1; };
        buildPhase = "runHook preBuild; runHook postBuild";
        checkPhase = ''
          runHook preCheck
          go vet ./...
          go test -race -covermode=atomic -coverprofile=coverage.out ./...
          ${pkgs.bash}/bin/bash ${./scripts/coverage-gate.sh} coverage.out
          runHook postCheck
        '';
        installPhase = ''
          runHook preInstall
          mkdir -p $out
          cp coverage.out $out/coverage.out
          runHook postInstall
        '';
      });

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
        default = chatcmsd;
      });

      devShells = forAllSystems (args: {
        default = mkDevShell args;
      });

      checks = forAllSystems (args: {
        go-build = mkServer args;
        go-test = mkGoTests args;
      });

      formatter = forAllSystems ({ pkgs, ... }: pkgs.nixpkgs-fmt);

      nixosModules.default = import ./nix/nixos-module.nix { inherit self; };
      nixosModules.chatcms = self.nixosModules.default;
    };
}
