# Attestation Report Verifizierungs-Webserver

Dieses Projekt implementiert einen Webserver zur Verifizierung von Attestation Reports. Die Berichte werden über einen HTTP-Endpunkt empfangen, verifiziert und in einer SQLite-Datenbank gespeichert.

## Features

- **Report-Verifizierung**: Verifiziert Attestation Reports mit dem `go-sev-guest/verify`-Paket.
- **SQLite-Integration**: Speichert die Berichte und deren Verifizierungsstatus.
- **HTTP-API**: Akzeptiert Berichte über den `/verify`-Endpunkt.
- **Sicherheit**: Schutz vor SQL-Injections und validierte Eingaben.

## Installation

1. Repository klonen.
2. Abhängigkeiten installieren:
   ```bash
   go mod tidy
   ```
3. Webserver starten:
   ```bash
   go run main.go
   ```

## API-Endpunkte

- `POST /verify`: Nimmt eine Datei entgegen und verifiziert den Attestation Report.

## Abhängigkeiten

- [go-sev-guest](https://github.com/google/go-sev-guest)
- [Gin Framework](https://github.com/gin-gonic/gin)
- [SQLite für Go](https://github.com/mattn/go-sqlite3)