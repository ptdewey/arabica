{ lib, buildGoModule, tailwindcss }:

buildGoModule rec {
  pname = "arabica";
  version = "0.1.0";
  src = ./.;
  vendorHash = "sha256-o7tm654ZGdPFyIZfhMCKi6QCb0re2PtQNJ3BM13vKzE=";

  nativeBuildInputs = [ tailwindcss ];

  preBuild = ''
    tailwindcss -i web/static/css/style.css -o web/static/css/output.css --minify
  '';

  buildPhase = ''
    runHook preBuild
    go build -o arabica cmd/server/main.go
    runHook postBuild
  '';

  installPhase = ''
    mkdir -p $out/bin
    mkdir -p $out/share/arabica

    # Copy static files, migrations, and templates
    cp -r web $out/share/arabica/
    cp -r migrations $out/share/arabica/
    cp -r internal $out/share/arabica/
    cp arabica $out/bin/arabica-unwrapped
    cat > $out/bin/arabica <<'EOF'
    #!/bin/sh
    SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
    SHARE_DIR="$SCRIPT_DIR/../share/arabica"
    cd "$SHARE_DIR"
    exec "$SCRIPT_DIR/arabica-unwrapped" "$@"
    EOF
    chmod +x $out/bin/arabica
  '';

  meta = with lib; {
    description = "Arabica - Coffee brew tracker";
    license = licenses.mit;
    platforms = platforms.linux;
    mainProgram = "arabica";
  };
}
