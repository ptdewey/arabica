# Arabica - Coffee Brew Tracker

A self-hosted web application for tracking your coffee brewing journey. Built with Go, Templ, and SQLite.

## Features

- 📝 Quick entry of brew data (temperature, time, method, flexible grind size entry, etc.)
- ☕ Organize beans by origin and roaster with quick-select dropdowns
- 📱 Mobile-first PWA design for on-the-go tracking
- 📊 Rating system and tasting notes
- 📥 Export your data as JSON
- 🔄 CRUD operations for all brew entries
- 🗄️ SQLite database with abstraction layer for easy migration

## Tech Stack

- **Backend**: Go 1.22+ (using stdlib router)
- **Database**: SQLite (via modernc.org/sqlite - pure Go, no CGO)
- **Templates**: Templ (type-safe HTML templates)
- **Frontend**: HTMX + Alpine.js
- **CSS**: Tailwind CSS
- **PWA**: Service Worker for offline support

## Project Structure

```
arabica/
├── cmd/server/          # Application entry point
├── internal/
│   ├── database/        # Database interface & SQLite implementation
│   ├── models/          # Data models
│   ├── handlers/        # HTTP handlers
│   └── templates/       # Templ templates
├── web/static/          # Static assets (CSS, JS, PWA files)
├── migrations/          # Database migrations
└── Makefile             # Build commands
```

## Getting Started

### Prerequisites

- Go 1.22+
- Templ CLI
- Tailwind CSS CLI
- (Optional) Air for hot reload

Or use Nix:

```bash
nix develop
```

### Installation

1. Clone the repository:
```bash
cd arabica-site
```

2. Install dependencies:
```bash
make install-deps
```

3. Build the application:
```bash
make build
```

4. Run the server:
```bash
make run
```

The application will be available at `http://localhost:8080`

### Development

For hot reload during development:

```bash
make dev
```

This uses Air to automatically rebuild when you change Go files or templates.

### Building Assets

```bash
# Generate templ files
make templ

# Build Tailwind CSS
make css

# Or build everything
make build
```

## Usage

### Adding a Brew

1. Navigate to "New Brew" from the home page
2. Select a bean (or add a new one with the "+ New" button)
   - When adding a new bean, provide a **Name** (required) like "Morning Blend" or "House Espresso"
   - Optionally add Origin, Roast Level, and Description
3. Select a roaster (or add a new one)
4. Fill in brewing details:
   - Method (Pour Over, French Press, etc.)
   - Temperature (°C)
   - Brew time (seconds)
   - Grind size (free text - enter numbers like "18" or "3.5" for grinder settings, or descriptions like "Medium" or "Fine")
   - Grinder (optional)
   - Tasting notes
   - Rating (1-10)
5. Click "Save Brew"

### Viewing Brews

Navigate to the "Brews" page to see all your entries in a table format with:
- Date
- Bean details
- Roaster
- Method and parameters
- Rating
- Actions (View, Delete)

### Exporting Data

Click "Export JSON" on the brews page to download all your data as JSON.

## Configuration

Environment variables:

- `DB_PATH`: Path to SQLite database (default: `./arabica.db`)
- `PORT`: Server port (default: `8080`)

## Database Abstraction

The application uses an interface-based approach for database operations, making it easy to swap SQLite for PostgreSQL or another database later. See `internal/database/store.go` for the interface definition.

## PWA Support

The application includes:
- Web App Manifest for "Add to Home Screen"
- Service Worker for offline caching
- Mobile-optimized UI with large touch targets

## Future Enhancements (Not in MVP)

- Statistics and analytics page
- CSV export
- Multi-user support (database already has user_id column)
- Search and filtering
- Photo uploads for beans/brews
- Brew recipes and sharing

## Development Notes

### Why These Choices?

- **Go**: Fast compilation, single binary deployment, excellent stdlib
- **modernc.org/sqlite**: Pure Go SQLite (no CGO), easy cross-compilation
- **Templ**: Type-safe templates, better than text/template for HTML
- **HTMX**: Progressive enhancement without heavy JS framework
- **Nix**: Reproducible builds across environments

### Database Schema

See `migrations/001_initial.sql` for the complete schema.

Key tables:
- `users`: Future multi-user support
- `beans`: Coffee bean information
- `roasters`: Roaster information
- `brews`: Individual brew records with all parameters

## License

MIT

## Contributing

This is a personal project, but suggestions and improvements are welcome!
