run:
    @LOG_LEVEL=debug LOG_FORMAT=console go run cmd/server/main.go

run-production:
    @LOG_FORMAT=json SECURE_COOKIES=true go run cmd/server/main.go

test:
    @go test ./... -cover -coverprofile=cover.out

style:
    @nix develop --command tailwindcss -i web/static/css/style.css -o web/static/css/output.css --minify
    # @tailwindcss -i web/static/css/style.css -o web/static/css/output.css --minify
