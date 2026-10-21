{ self }:
{ config, lib, pkgs, ... }:

let
  cfg = config.services.chatcms;
in
{
  options.services.chatcms = {
    enable = lib.mkEnableOption "Quiqr ChatCMS server";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.chatcmsd;
      defaultText = lib.literalExpression "self.packages.\${system}.chatcmsd";
      description = "The chatcmsd package to run.";
    };

    listenAddress = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1";
      description = "Address chatcmsd binds to.";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 8080;
      description = "Port chatcmsd listens on.";
    };

    baseUrl = lib.mkOption {
      type = lib.types.str;
      example = "https://chatcms.eu";
      description = "Public base URL, used in invitation links and previews.";
    };

    stateDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/chatcms";
      description = "Directory holding project working copies, media caches and build output.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "chatcms";
      description = "User the service runs as.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "chatcms";
      description = "Group the service runs as.";
    };

    database = {
      createLocally = lib.mkOption {
        type = lib.types.bool;
        default = true;
        description = "Provision a local PostgreSQL database and role for ChatCMS.";
      };

      name = lib.mkOption {
        type = lib.types.str;
        default = "chatcms";
        description = "PostgreSQL database name.";
      };

      urlFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = ''
          File containing the PostgreSQL connection string. Required when
          database.createLocally is false. Read at start-up, never put in the
          Nix store.
        '';
      };
    };

    secretKeyFile = lib.mkOption {
      type = lib.types.path;
      description = ''
        File containing the master key used to encrypt per-project SSH deploy
        keys and session secrets at rest. Never put in the Nix store.
      '';
    };

    hugoPackage = lib.mkOption {
      type = lib.types.package;
      default = pkgs.hugo;
      description = "Hugo used for builds and previews.";
    };

    gitPackage = lib.mkOption {
      type = lib.types.package;
      default = pkgs.git;
      description = "Git CLI used for all repository operations.";
    };

    extraEnvironment = lib.mkOption {
      type = lib.types.attrsOf lib.types.str;
      default = { };
      description = "Extra environment variables for the service.";
    };
  };

  config = lib.mkIf cfg.enable {
    assertions = [
      {
        assertion = cfg.database.createLocally || cfg.database.urlFile != null;
        message = "services.chatcms: set database.urlFile when database.createLocally is false.";
      }
    ];

    users.users = lib.mkIf (cfg.user == "chatcms") {
      chatcms = {
        isSystemUser = true;
        group = cfg.group;
        home = cfg.stateDir;
      };
    };

    users.groups = lib.mkIf (cfg.group == "chatcms") {
      chatcms = { };
    };

    services.postgresql = lib.mkIf cfg.database.createLocally {
      enable = true;
      ensureDatabases = [ cfg.database.name ];
      ensureUsers = [
        {
          name = cfg.user;
          ensureDBOwnership = true;
        }
      ];
    };

    systemd.services.chatcms = {
      description = "Quiqr ChatCMS server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network-online.target" ] ++ lib.optional cfg.database.createLocally "postgresql.service";
      wants = [ "network-online.target" ];
      requires = lib.optional cfg.database.createLocally "postgresql.service";

      path = [ cfg.gitPackage cfg.hugoPackage pkgs.openssh ];

      environment = {
        CHATCMS_LISTEN = "${cfg.listenAddress}:${toString cfg.port}";
        CHATCMS_BASE_URL = cfg.baseUrl;
        CHATCMS_STATE_DIR = cfg.stateDir;
        CHATCMS_GIT_BIN = lib.getExe cfg.gitPackage;
        CHATCMS_HUGO_BIN = lib.getExe cfg.hugoPackage;
        CHATCMS_SECRET_KEY_FILE = cfg.secretKeyFile;
      } // lib.optionalAttrs cfg.database.createLocally {
        CHATCMS_DATABASE_URL = "postgres:///${cfg.database.name}?host=/run/postgresql";
      } // lib.optionalAttrs (cfg.database.urlFile != null) {
        CHATCMS_DATABASE_URL_FILE = cfg.database.urlFile;
      } // cfg.extraEnvironment;

      serviceConfig = {
        ExecStart = "${lib.getExe cfg.package}";
        User = cfg.user;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = 5;
        StateDirectory = lib.mkIf (cfg.stateDir == "/var/lib/chatcms") "chatcms";
        WorkingDirectory = cfg.stateDir;
        ReadWritePaths = [ cfg.stateDir ];
        AmbientCapabilities = [ ];
        NoNewPrivileges = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [ "AF_UNIX" "AF_INET" "AF_INET6" ];
        RestrictNamespaces = false;
        LockPersonality = true;
        SystemCallArchitectures = "native";
      };
    };
  };
}
