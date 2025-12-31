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

          nativeBuildInputs = with pkgs; [ templ tailwindcss ];

          preBuild = ''
            # Generate templates before building
            templ generate
            
            # Build Tailwind CSS
            tailwindcss -i web/static/css/style.css -o web/static/css/output.css --minify
          '';

          # Build output goes to bin/arabica
          buildPhase = ''
            runHook preBuild
            go build -o arabica cmd/server/main.go
            runHook postBuild
          '';

          installPhase = ''
            mkdir -p $out/bin
            mkdir -p $out/share/arabica

            # Copy static files and migrations
            cp -r web $out/share/arabica/
            cp -r migrations $out/share/arabica/

            # Install the actual binary
            cp arabica $out/bin/arabica-unwrapped

            # Create wrapper script that changes to the share directory
            cat > $out/bin/arabica <<'EOF'
            #!/bin/sh
            SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
            SHARE_DIR="$SCRIPT_DIR/../share/arabica"
            cd "$SHARE_DIR"
            exec "$SCRIPT_DIR/arabica-unwrapped" "$@"
            EOF
            chmod +x $out/bin/arabica
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
