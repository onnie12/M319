# Codera Battle

Rundenbasiertes Kommandozeilen-Kampfspiel in Go: Sechs Helden treten gegen den
Endgegner «Entropie-Drache» an. Abschlussprojekt der Module **M319**
(Applikation realisieren) und **M164** (Datenbank erstellen und Daten einfügen).

## Voraussetzungen

- Go 1.22 oder neuer
- Docker (für die PostgreSQL-Datenbank)

## Setup

### 1. Datenbank starten (PostgreSQL via Docker)

```bash
docker volume create codera_pgdata
docker run -d \
  --name codera_postgres \
  -e POSTGRES_USER=codera \
  -e POSTGRES_PASSWORD=codera \
  -e POSTGRES_DB=codera \
  -p 5433:5432 \
  -v codera_pgdata:/var/lib/postgresql/data \
  --restart unless-stopped \
  postgres:16
```

### 2. Konfiguration anlegen

```bash
cp .env-example .env
```

Die Werte in `.env` bei Bedarf an die eigene Datenbank anpassen (Port, Benutzer,
Passwort). `.env` ist per `.gitignore` ausgeschlossen und wird nicht committet.

### 3. Spiel starten

```bash
go run .
```

Beim Start verbindet sich das Spiel mit der Datenbank, legt die Tabellen an
(`Migrate`), befüllt sie einmalig (`Seed`, idempotent) und lädt die Helden.
Ist keine Datenbank erreichbar, fällt das Spiel automatisch auf Standarddaten
zurück und läuft trotzdem.

## Tests

```bash
go test ./...
```

Die Datenbank-Tests verwenden eine In-Memory-SQLite und benötigen **kein**
laufendes PostgreSQL.

## Bedienung

Pro Runde wählt man für jeden Helden eine Aktion per Nummer; mit «Enter» führt
der Held automatisch eine sinnvolle Aktion aus. Jede Aktion wird zusätzlich
strukturiert in `logs/battle-JJJJ-MM-TT.log` protokolliert (getrennt von der
Bildschirmausgabe).

## Projektstruktur

```
.
├── main.go        # Zusammenbau: .env, Logging, Helden laden (DB mit Fallback), Kampfstart
├── internal/      # Gemeinsamer Contract: Combatant, HeroController, Loadout, ActionContext ...
├── dragon/        # Entropie-Drache inkl. KI und Rage-Modus (vorgegeben)
├── hero/<rolle>/  # Je eine Heldenklasse; erfüllt das HeroController-Interface
├── combat/        # Kampf-Loop, Initiative, Zugabwicklung, Schadensformel
├── db/            # PostgreSQL-Persistenz (GORM): Modelle, Verbindung, Seed, Laden
└── logging/       # Strukturiertes Logging (log/slog + Rotation via lumberjack)
```

## Architektur (Kurzüberblick)

- Jede Heldenklasse erfüllt das Interface `internal.HeroController` in ihrem
  **eigenen** Paket.
- `combat` und `db` importieren **nie** ein `hero/*`-Paket. Nur `main` kennt die
  konkreten Helden und ordnet Rolle → Konstruktor über eine kleine Registry zu.
  Das hält die Pakete entkoppelt, vermeidet Import-Zyklen und trennt die
  Verantwortlichkeiten sauber.
- `db` liefert reine Daten (`[]internal.Loadout`); `main` baut daraus die Helden.

## Team

| Rolle | Person |
|-------|--------|
| Funktions-Krieger | Onni Johansson |
| Arkan-Dokumentar | Roda Ikwueto |
| Daten-Druide | Jonas Aeschlimann |
| System-Infiltrator | Luca Witkowski |

Die zusätzlichen Klassen Code-Kleriker und Runenschmied wurden zur
Vervollständigung der sechs Rollen ergänzt.
