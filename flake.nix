{
  description = "Dev Shells Flake";
  inputs = { nixpkgs.url = "nixpkgs/nixpkgs-unstable"; };
  outputs = { nixpkgs, self, ... }:
    let
      forAllSystems = function:
        nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ]
        (system: function nixpkgs.legacyPackages.${system} system);
    in {
      devShells = forAllSystems (pkgs: system: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            templ
            tailwindcss
            air # Live reload for Go
          ];
        };
      });

      packages = forAllSystems (pkgs: system: rec {
        arabica = pkgs.buildGoModule {
          pname = "arabica";
          version = "0.1.0";
          src = ./.;

          # Vendor hash for Go dependencies
          vendorHash = "sha256-7QYmui8+jyG/QOds0YfZfgsKqZcvm/RLQCkDFUk+xUc=";

          nativeBuildInputs = with pkgs; [ templ ];

          preBuild = ''
            # Generate templates before building
            templ generate
          '';

          # Build output goes to bin/arabica
          buildPhase = ''
            runHook preBuild
            go build -o arabica cmd/server/main.go
            runHook postBuild
          '';

          installPhase = ''
            mkdir -p $out/bin
            cp arabica $out/bin/

            # Copy static files and migrations
            mkdir -p $out/share/arabica
            cp -r web $out/share/arabica/
            cp -r migrations $out/share/arabica/
          '';

          meta = with pkgs.lib; {
            description = "Arabica - Coffee brew tracker";
            license = licenses.mit;
            platforms = platforms.linux;
          };
        };

        default = arabica;
      });

      apps = forAllSystems (pkgs: system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.arabica}/bin/arabica";
        };
      });
    };
}
