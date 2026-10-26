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

    firstAdministrator = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "pim@example.org";
      description = ''
        Email address of the person who administers this installation. They
        are made an administrator every time the service starts, which is how
        a new installation gets its first one: nobody is there to appoint
        anybody. The account must already exist; until it does, the service
        starts and says it could not appoint anybody.
      '';
    };

    mail = {
      host = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        example = "smtp.example.org";
        description = ''
          Mail server invitations and password resets are sent through. With
          none, those links go to the service log instead, which suits an
          installation run by the person who owns it and nothing larger.
        '';
      };

      port = lib.mkOption {
        type = lib.types.port;
        default = 587;
        description = "Port on the mail server.";
      };

      from = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        example = "chatcms@example.org";
        description = ''
          Address messages come from. Defaults to chatcms@ the host of
          baseUrl, which is what a receiving server is most likely to accept.
        '';
      };

      fromName = lib.mkOption {
        type = lib.types.str;
        default = "Quiqr ChatCMS";
        description = "Name a reader sees instead of the address.";
      };

      username = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Username for the mail server, when it wants one.";
      };

      passwordFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        description = ''
          File holding the mail server password. Never put in the Nix store;
          systemd reads it and hands the service a copy only it can read.
        '';
      };

      startTLS = lib.mkOption {
        type = lib.types.bool;
        default = true;
        description = ''
          Ask for the connection to be encrypted. Leave it on for anything
          reachable over a network.
        '';
      };
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

    bubblewrapPackage = lib.mkOption {
      type = lib.types.package;
      default = pkgs.bubblewrap;
      description = ''
        Bubblewrap, which isolates every site build. Without it, builds are
        refused rather than run unisolated.
      '';
    };

    utilLinuxPackage = lib.mkOption {
      type = lib.types.package;
      default = pkgs.util-linux;
      description = "Provides prlimit, which bounds the memory a build may use.";
    };

    buildTimeout = lib.mkOption {
      type = lib.types.str;
      default = "10m";
      example = "2m";
      description = "How long a site build may run before it is ended.";
    };

    workflowTools = lib.mkOption {
      type = lib.types.listOf lib.types.package;
      default = with pkgs; [ bash coreutils gnugrep gnused findutils diffutils ];
      description = ''
        What a site's own check step may use. Fixed and pinned: a workflow
        cannot ask for another tool, because fetching software during a
        publish is exactly what the sandbox exists to prevent.
      '';
    };

    checkTimeout = lib.mkOption {
      type = lib.types.str;
      default = "5m";
      description = "How long a site's own check may run before it is ended.";
    };

    checkMemory = lib.mkOption {
      type = lib.types.str;
      default = "1GiB";
      description = "How much memory a site's own check may use.";
    };

    buildMemory = lib.mkOption {
      type = lib.types.str;
      default = "2GiB";
      example = "512MiB";
      description = ''
        How much memory a site build may use. A build that exceeds it is
        ended and its log is kept.
      '';
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

      path = [
        cfg.gitPackage
        cfg.hugoPackage
        cfg.bubblewrapPackage
        cfg.utilLinuxPackage
        pkgs.openssh
      ];

      environment = {
        CHATCMS_LISTEN = "${cfg.listenAddress}:${toString cfg.port}";
        CHATCMS_BASE_URL = cfg.baseUrl;
        CHATCMS_STATE_DIR = cfg.stateDir;
        CHATCMS_GIT_BIN = lib.getExe cfg.gitPackage;
        CHATCMS_HUGO_BIN = lib.getExe cfg.hugoPackage;
        # systemd reads the key as root and hands the service a copy only it
        # can read, so the key on disk stays owned by root and the service
        # does not run as root to read it. %d is where systemd puts it.
        CHATCMS_SECRET_KEY_FILE = "%d/master-key";
        CHATCMS_BWRAP_BIN = "${cfg.bubblewrapPackage}/bin/bwrap";
        CHATCMS_PRLIMIT_BIN = "${cfg.utilLinuxPackage}/bin/prlimit";
        CHATCMS_BUILD_TIMEOUT = cfg.buildTimeout;
        CHATCMS_BUILD_MEMORY = cfg.buildMemory;
        CHATCMS_WORKFLOW_TOOLS =
          lib.concatStringsSep ":" (map (tool: "${tool}/bin") cfg.workflowTools);
        CHATCMS_CHECK_TIMEOUT = cfg.checkTimeout;
        CHATCMS_CHECK_MEMORY = cfg.checkMemory;
      } // lib.optionalAttrs (cfg.firstAdministrator != null) {
        CHATCMS_FIRST_ADMINISTRATOR = cfg.firstAdministrator;
      } // lib.optionalAttrs (cfg.mail.host != null) {
        CHATCMS_MAIL_HOST = cfg.mail.host;
        CHATCMS_MAIL_PORT = toString cfg.mail.port;
        CHATCMS_MAIL_FROM_NAME = cfg.mail.fromName;
        CHATCMS_MAIL_STARTTLS = if cfg.mail.startTLS then "yes" else "no";
      } // lib.optionalAttrs (cfg.mail.from != null) {
        CHATCMS_MAIL_FROM = cfg.mail.from;
      } // lib.optionalAttrs (cfg.mail.username != null) {
        CHATCMS_MAIL_USERNAME = cfg.mail.username;
      } // lib.optionalAttrs (cfg.mail.passwordFile != null) {
        CHATCMS_MAIL_PASSWORD_FILE = "%d/mail-password";
      } // lib.optionalAttrs cfg.database.createLocally {
        CHATCMS_DATABASE_URL = "postgres:///${cfg.database.name}?host=/run/postgresql";
      } // lib.optionalAttrs (cfg.database.urlFile != null) {
        CHATCMS_DATABASE_URL_FILE = cfg.database.urlFile;
      } // cfg.extraEnvironment;

      serviceConfig = {
        ExecStart = "${lib.getExe cfg.package}";
        LoadCredential = [ "master-key:${cfg.secretKeyFile}" ]
          ++ lib.optional (cfg.mail.passwordFile != null) "mail-password:${cfg.mail.passwordFile}";
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
