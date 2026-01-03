{ config, lib, pkgs, ... }:

let cfg = config.services.arabica;
in {
  options.services.arabica = {
    enable = lib.mkEnableOption "Arabica coffee brew tracking service";

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.callPackage ./default.nix { };
      defaultText = lib.literalExpression "pkgs.callPackage ./default.nix { }";
      description = "The arabica package to use.";
    };

    settings = {
      port = lib.mkOption {
        type = lib.types.port;
        default = 18910;
        description = "Port on which the arabica server listens.";
      };
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/arabica";
      description = "Directory where arabica stores its database.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "arabica";
      description = "User account under which arabica runs.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "arabica";
      description = "Group under which arabica runs.";
    };

    openFirewall = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Whether to open the firewall for the arabica port.";
    };
  };

  config = lib.mkIf cfg.enable {
    users.users.${cfg.user} = lib.mkIf (cfg.user == "arabica") {
      isSystemUser = true;
      group = cfg.group;
      description = "Arabica service user";
      home = cfg.dataDir;
      createHome = true;
    };

    users.groups.${cfg.group} = lib.mkIf (cfg.group == "arabica") { };

    systemd.services.arabica = {
      description = "Arabica Coffee Brew Tracking Service";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        Type = "simple";
        User = cfg.user;
        Group = cfg.group;
        ExecStart = "${cfg.package}/bin/arabica";
        Restart = "on-failure";
        RestartSec = "10s";

        # Security hardening
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [ cfg.dataDir ];
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" ];
        RestrictNamespaces = true;
        LockPersonality = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        MemoryDenyWriteExecute = true;
        SystemCallArchitectures = "native";
        CapabilityBoundingSet = "";
      };

      environment = {
        PORT = toString cfg.settings.port;
        DB_PATH = "${cfg.dataDir}/arabica.db";
      };
    };

    networking.firewall =
      lib.mkIf cfg.openFirewall { allowedTCPPorts = [ cfg.settings.port ]; };
  };
}
