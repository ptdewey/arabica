run:
    @go run cmd/server/main.go

style:
    @tailwindcss -i web/static/css/style.css -o web/static/css/output.css --minify


