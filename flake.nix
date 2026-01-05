{
  description = "Arabica - Coffee brew tracking application";
  inputs = { nixpkgs.url = "nixpkgs/nixpkgs-unstable"; };
  outputs = { nixpkgs, self, ... }:
    let
      forAllSystems = function:
        nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ]
        (system: function nixpkgs.legacyPackages.${system} system);
    in {
      devShells = forAllSystems (pkgs: system: {
        default = pkgs.mkShell { packages = with pkgs; [ go tailwindcss ]; };
      });

      packages = forAllSystems (pkgs: system: rec {
        arabica = pkgs.callPackage ./default.nix { };
        default = arabica;
      });

      apps = forAllSystems (pkgs: system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.arabica}/bin/arabica";
        };
      });

      nixosModules.default = import ./module.nix;
    };
}
