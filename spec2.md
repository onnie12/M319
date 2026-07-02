Auftrag M164 – Datenbank-Infrastruktur für
den Kampf gegen den Entropie-Drachen
Story-Einstieg
Der Entropiedrache wartet. Eure Gruppe hat die Runen-Artefakte geschmiedet und die KampfLogik in Go-Code gegossen – doch ohne die Datenbank der Welt sind eure Helden nur blosse
Schatten. Die Entity hat euch den Auftrag erteilt, das Fundament zu legen: Eine PostgreSQLDatenbank, die die Helden, ihre Ausrüstung und ihre Fähigkeiten in der realen Welt verankert.
Der Arkan-Dokumentar*in hat die Blaupause gezeichnet – das logisch relationale Modell eurer
Helden. Nun müsst ihr dieses Modell in die Realität umsetzen. Docker Desktop wird euer
Werkzeug sein, um die Datenbank-Instanz zu erschaffen. Euer Vorgehen wird zur heiligen
Schrift, damit jeder in der Gruppe die gleiche Umgebung aufsetzen kann. Denn nur wenn alle
die gleiche Datenbank haben, können eure Helden gemeinsam gegen den Entropie-Drachen
antreten.
Die Entity spricht: „Die Runen sind initialisiert, der Code ist bereit. Doch ohne den Speicher der
Welt – die Datenbank – verpufft eure Magie im Nichts. Erschafft die Instanz. Füllt sie mit Leben.
Dokumentiert jeden Schritt. Die Welt von Codera ruht auf euren Tabellen."
Arbeitsauftrag
Jeder der nachfolgenden Punkte muss nachweisbar erfüllt sein (via DDL-"Code" oder in der
Gruppendokumentation).
Dieser Auftrag läuft parallel zum M319-Auftrag (Programmieren). Während der M319-Auftrag den KampfCode implementiert, stellt dieser M164-Auftrag die Datenbank-Infrastruktur bereit, die der Code
benötigt. Die Aufgaben sind so aufgeteilt, dass sie möglichst gleichzeitig bearbeitet werden können:
Person M319 (Programmieren) M164 (Datenbanken)
ArkanDokumentar*in
C4-Diagramme, Clean Code,
Linter
Dokumentation des DB-Setups prüfen &
abnehmen
Daten-Druide
GORM-Modelle, DBConnection-Code
DB aufsetzen, Vorgehen dokumentieren
Person M319 (Programmieren) M164 (Datenbanken)
Code-Kleriker*in Logging-Framework DB-Zugangsdaten via .env verteilen
FunktionsKrieger*in
Kampf-Loop, Held interagiert
mit DB
DB-Verbindung testen, Query-Ergebnisse
validieren
Runenschmied*in
DB-Migration & Seeds im GoCode
SQL-DDL & Seeds parallel als SQL-Skript
vorbereiten
Bewertungsraster
Datenbank-Setup (Docker)
Docker Desktop installiert und funktionsfähig
PostgreSQL-Container gestartet und läuft
Datenbank codera wurde erstellt
Verbindung zur Datenbank ist von localhost möglich
Datenbanksetup und Beweisscreenshot das alles funktioniert in der Gruppendokumentation
Schema-Umsetzung (DDL)
ERD selbst gezeichnet (aus der Problembeschreibung unten abgeleitet)
Alle Tabellen wurden erstellt (Helden, Equipment, Skill)
Primär- und Fremdschlüssel korrekt definiert
Datentypen sinnvoll gewählt
Constraints (NOT NULL, UNIQUE, Defaults) wurden angewandt
Seed-Daten (DML)
Helden wurden mit korrekten Namen in die DB eingefügt
Ausrüstungsgegenstände wurden erfasst
Fähigkeiten (Skills) wurden den Rollen zugeordnet
Daten sind konsistent (referenzielle Integrität)
Daten abfragen & prüfen
SELECT-Abfragen zeigen alle Helden an
WHERE-Filter und JOIN-Abfragen funktionieren
Export der Datenbank als SQL-Dump
Import des SQL-Dumps in eine frische DB
Dokumentation
Schritt-für-Schritt-Anleitung (Markdown) für das gesamte Setup
Enthält: Docker-Befehle, DDL-SQL, Seed-SQL, Export/Import-Befehle
Enthält: Selbst gezeichnetes ERD (abgeleitet aus der Problembeschreibung)
Enthält: Beispiel-Queries mit Ergebnissen
Enthält: Troubleshooting-Sektion (häufige Fehler & Lösungen)
Bonus (Docker Compose)
docker-compose.yml erstellt (PostgreSQL-Service + Volume)
.env -Variablen werden von Compose verwendet
docker compose up startet die DB vollständig automatisiert
Docker-Compose-Konfiguration ist in der Dokumentation nachvollziehbar beschrieben
Detaillierte Aufgaben
Aufgabe 1: Docker-Infrastruktur planen und aufsetzen (Lernziele B1G,
B1F)
Eine Person aus der Gruppe (z. B. der Daten-Druide) setzt die Datenbank-Infrastruktur auf und
dokumentiert dabei jeden Schritt so genau, dass die anderen Gruppenmitglieder ihn 1:1 reproduzieren
können.
Schritte:
1. Docker Desktop installieren (falls nicht vorhanden)
2. PostgreSQL-Image pullen
3. Container erstellen und starten (Port 5432, Volume für Persistenz)
4. Datenbank codera erstellen
5. User codera mit Passwort codera einrichten und Berechtigungen vergeben
6. Verbindung testen (mit pgadmin4)
Dokumentation: Schreiben Sie dabei jedes Kommando auf, das ausgeführt wird. Erklären Sie kurz, was
es bewirkt. Notieren Sie auch Fehler, die aufgetreten sind, und wie sie ausgelöst wurden.
Für die anderen: Nachdem die Dokumentation fertig ist, setzen alle anderen Gruppenmitglieder die
Datenbank exakt gleich auf. Erst wenn alle eine funktionierende DB-Instanz haben, kann der M319-Code
bei allen laufen.
Wichtig: Abgesehen vom initialen Setup (Dokumentation der Docker-Infrastruktur) arbeitet jede Person
auf ihrer eigenen PostgreSQL-Instanz. Das heisst: Sobald die Setup-Dokumentation fertig ist, erstellt
jedes Gruppenmitglied seinen eigenen Container. Alle weiteren Aufgaben (DDL, Seeds, Queries,
Export/Import) werden von jedem Gruppenmitglied auf der persönlichen Datenbank ausgeführt – nicht
auf einer gemeinsamen Instanz.
Aufgabe 2: ERD aus der Problembeschreibung ableiten und zeichnen
(Lernziele A1G, A1F, B1E, B2E, B3E)
Die Entity hat euch den Auftrag erteilt, das Datenfundament für den Kampf gegen den EntropieDrachen zu legen. In der folgenden Beschreibung erfahrt ihr, welche Daten benötigt werden. Leitet
daraus selbständig ein logisch relationales ERD ab – zeichnet es, diskutiert es in der Gruppe, und lasst
es vom Arkan-Dokumentar*in abnehmen, bevor ihr es als DDL umsetzt.
Problembeschreibung – Die Helden-Datenbank von Codera
In Codera kämpfen Helden gegen den Entropie-Drachen. Jeder Held wird durch eine reale Person
gespielt und hat einen Namen, eine Rolle (z. B. Arkan-Dokumentar*in, Daten-Druide, Code-Kleriker*in,
Funktions-Krieger*in, Runenschmied*in, System-Infiltrator*in) sowie Kampfwerte (Abschliessende
Liste: maximale Trefferpunkte, aktuelle Trefferpunkte, Angriff, Verteidigung, Geschwindigkeit).
Helden können Ausrüstungsgegenstände tragen – maximal drei gleichzeitig: eine Waffe, eine Rüstung
und ein Accessoire. Jeder Gegenstand hat einen Namen, einen Typ (Auswahl aus:
weapon/armor/accessory) und kann verschiedene Boni gewähren (Auswahl aus: Angriffsbonus,
Verteidigungsbonus, Geschwindigkeitsbonus, Trefferpunktbonus). Ein Ausrüstungsgegenstand kann
dabei einen oder mehrere Boni haben. Es gibt auch einen Bonus-Effekt für besondere Fähigkeiten
(siehe nächsten Abschnitt). Ein Gegenstand kann von mehreren Helden getragen werden, aber nicht
jeder Held trägt zwingend alle drei Gegenstände.
Jeder Held beherrscht ausserdem Fähigkeiten (Skills), die an seine Rolle gebunden sind. Zum Beispiel
hat ein Funktions-Krieger andere Skills als ein Daten-Druide. Ein Skill hat einen Namen, eine
Beschreibung, Schadensminimum und -maximum (beides 0 bei reinen Heil-Skills), einen Heilungswert
(der bei einem Kampfskill 0 ist), eine Genauigkeit (0.0–1.0) und einen Zieltyp (z. B. einzelner Gegner, alle
Gegner, einzelner Verbündeter, alle Verbündete, selbst). Ein Skill gehört immer zu genau einer Rolle.
Also der Skill kann nur von genau der Rolle "gelernt" bzw. "genutzt" werden.
Ein Held kann seine Rolle nicht wechseln – die Rolle bestimmt, welche Skills verfügbar sind.
Aufgaben:
1. ERD zeichnen (A1G, A1F): Leiten Sie aus der Problembeschreibung die Entitäten, Attribute,
Beziehungen und Kardinalitäten ab. Zeichnen Sie das dazugehörige ERD (digital mit draw.io) . Das
fertige ERD muss in der Dokumentation enthalten sein.
2. ERD erläutern (A1G): Beschreiben Sie in der Dokumentation mit eigenen Worten, was die
einzelnen Entitäten, Attribute und Beziehungen bedeuten. Zeige Sie die Kardinalitäten auf.
3. ERD kritisch hinterfragen (A1E): Diskutieren Sie in der Gruppe, ob das ERD vollständig ist. Fehlt
etwas? Gibt es Redundanzen? Wie könnte man es verbessern? Halten Sie Ihre Überlegungen in der
Dokumentation fest (Der Verbesserungsprozess des ERD soll sichtbar sein!).
4. DDL erstellen (B1E, B2E, B3E): Schreiben Sie ein SQL-Skript, das alle Tabellen mit passenden
Datentypen, Constraints, Primär- und Fremdschlüsseln erstellt (genau so wie Sie es im ERD
definiert haben!). Verwenden Sie dabei SERIAL oder IDENTITY für Auto-Increment.
5. Schema umsetzen (B1F): Führen Sie das DDL-Skript gegen eine der PostgreSQL-Instanzen aus.
Halten Sie das Vorgehen und das Ergebnis in der Dokumentation fest.
Hinweis: Der M319-Code nutzt GORM AutoMigrate, um die Tabellen zu erstellen. Für diesen
Auftrag erstellen Sie die Tabellen jedoch manuell mit DDL. Vergleichen Sie anschliessend das
von Ihnen selbst gebaute Schema mit dem von GORM generierten – das hilft Ihnen,
Abweichungen zu erkennen und zu verstehen. Halten Sie diese Unterschiede ebenfalls in der
Dokumentation fest!
Aufgabe 3: Seed-Daten einfügen (Lernziele C1G, C1E, C2F, C3E)
Fügen Sie nun Daten in die Tabellen ein, damit der M319-Code Helden laden könnte.
Vorgaben für die Seed-Daten:
Jedes Gruppenmitglied erstellt seinen eigenen Helden mit seinem/ihrem echten Namen (oder
dem Heldennamen den Sie zu Beginn des Moduls festgelegt haben)
Jeder Held hat 3 Skills (passend zur Rolle)
Jeder Held hat 3 Ausrüstungsgegenstände (Waffe, Rüstung, Accessoire)
Referenz-Daten: Die folgenden Basisdaten dienen als Orientierung. Passen Sie sie auf Ihre Gruppe an.
Helden:
Name Rolle MaxHP Attack Defense Speed
<DEIN_NAME> arkan 120 18 8 14
<DEIN_NAME> druide 100 14 10 16
<DEIN_NAME> kleriker 110 10 12 12
<DEIN_NAME> krieger 150 22 14 8
<DEIN_NAME> schmied 130 16 16 10
<DEIN_NAME> infiltrator 120 30 10 20
Ausrüstung:
Rolle Waffe Rüstung Accessoire
arkan Pergament-Stab (ATK+8)
Runen-Gewand
(DEF+5)
Tintenfass-Amulett (SPD+3, HP+20)
druide
Transformations-Kristall
(ATK+6)
Datenstrom-Mantel
(DEF+4)
Schema-Ring (SPD+5, HP+10)
kleriker Debugger-Stab (ATK+4) Kleriker-Robe (DEF+6)
Auge-des-Debuggers-Amulett
(SPD+2, HP+30)
krieger
Funktions-Schwert
(ATK+10)
Krieger-Rüstung
(DEF+8)
Gurt-der-Ausdauer (SPD+2, HP+40)
schmied
Architekten-Hammer
(ATK+7)
Runen-Plattenpanzer
(DEF+9)
Siegelring-der-Stabilität (SPD+1,
HP+25)
infiltrator Schatten-Dolch (ATK+14)
Infiltrator-Cape
(DEF+5)
Amulett-der-Verwundbarkeit
(SPD+5, HP+25)
Skills pro Rolle:
Rolle Skill 1 Skill 2 Skill 3
arkan
Runen-Geschoss (12-24
DMG, 90%)
Arkaner Bann (8-16 DMG,
85%, AOE)
Klärende-Annotation (Heal
20)
druide
Datenklinge (10-20 DMG,
85%)
Strukturwandel (14-28 DMG,
70%)
TransformativeRegeneration (Heal 16, self)
kleriker
Heiliges-Licht (6-12 DMG,
95%)
Heilsame-Korrektur (Heal 27,
single)
Segen-der-Stabilität (Heal 12,
all)
krieger
Präziser-Hieb (18-32
DMG, 80%)
Schutzschild (+5 DEF self)
Kampfschrei (8-16 DMG, 90%,
ATK-Buff)
schmied
Architekten-Schlag (14-
26 DMG, 85%)
Schutz-Rune (+3 DEF all)
Konstrukt-Schild (-50% DMG
single)
infiltrator
Hinterhalt (22-40 DMG,
80%)
Schwachstelle-analysieren
(-5 DEF dragon)
Tödliche-Praezision (18-34
DMG, 90%)
Aufgaben:
1. INSERT-Skript (C1E): Schreiben Sie die SQL-INSERTs für Ausrüstung, Skills und Helden. Beachten
Sie die Reihenfolge: zuerst Equipment (unabhängig), dann Skills (unabhängig), dann Helden mit
den korrekten equipped_*-FKs.
2. Daten-Integrität (C3E): Achten Sie auf referenzielle Integrität – verwenden Sie INSERT in der
richtigen Reihenfolge, damit Fremdschlüssel zu jedem Zeitpunkt gültig sind.
3. Bulk-Import (C2F): Erstellen Sie Ihre Seed Daten in einem CSV File und importieren Sie diese via
CSV Import in die Datenbank.
Dokumentieren Sie die Lösung aller Teilaufgaben und das dazugehörige Resultat in der Dokumentation.
Aufgabe 4: Daten abfragen und prüfen (Lernziele D1G, D1F)
Führen Sie die folgenden Abfragen aus und dokumentieren Sie die Ergebnisse.
1. Alle Helden anzeigen – Zeige alle Helden mit Namen und Rolle.
2. Ausrüstung eines Helden – Zeige für einen bestimmten Helden alle Ausrüstungsgegenstände (mit
JOIN).
3. Skills einer Rolle – Zeige alle Skills einer spezifischen Rolle (z.B. „krieger".)
4. Helden mit voller Ausrüstung – Finde alle Helden, die alle 3 Ausrüstungsslots belegt haben.
5. Durchschnittlicher Angriff pro Rolle – Berechne den durchschnittlichen Attack-Wert gruppiert
nach Rolle.
Dokumentieren Sie alle Abfragen und das dazugehörige Resultat in der Dokumentation.
Aufgabe 5: Datenbank exportieren und importieren (Lernziele C2G)
1. Export: Erstelle Sie einen SQL-Dump (Backup) der Datenbank.
2. Import: Löschen Sie die Datenbank, erstelle Sie sie neu und importiere den Dump.
3. Verifikation: Führen Sie die SELECT-Abfragen aus Aufgabe 4 erneut aus und bestätigen Sie, dass
alle Daten noch identisch sind.
4. Dokumentation: Halten Sie das genaue Vorgehen und das Resultat in der Dokumentation fest.
Aufgabe 6 (Bonus): Docker Compose (Bonusziel für Bonuspunkte)
Erstellen Sie eine docker-compose.yml , die den PostgreSQL-Container vollständig automatisiert
aufsetzt. Testen Sie bei einem der anderen Lernenden Ihrer Gruppe ob das automatische Setup wirklich
funktioniert. Dokumentieren Sie das Vorgehen und das Resultat in der Dokumentation und legen Sie das
docker-compose.yml File in das Git Repo des M319 Auftrags.
Anforderungen:
PostgreSQL als Docker Container
Persistente Daten über Volume
Umgebungsvariablen aus .env -Datei
Port 5432 wird exponiert
Der Container startet (fehlerfrei) mit docker compose up
Liefergegenstände (Deliverables)
dokumentation_gruppe_x.pdf – Gruppendokumentation, die die Lösung und die
Resultate aller Aufträge enthält
ddl.sql – DDL-Skript (CREATE TABLEs)
seed.sql – Seed-Daten (INSERTs)
queries.sql – SELECT-Abfragen mit Kommentaren
export.sql – pg_dump (Backup) des fertigen DB-Zustands
docker-compose.yml (Optional --> Für Bonuspunkte)
.env-example (Vorlage für DB-Zugangsdaten)
Glossar
Begriff Erklärung
Container
Hier: Ein Docker-Container – eine gekapselte Ausführungsumgebung für die
PostgreSQL-Datenbank.
Docker Eine Plattform zum Ausführen von Anwendungen in isolierten Containern.
DDL
Data Definition Language – SQL-Befehle zum Erstellen und Verändern von
Datenbank-Strukturen (CREATE TABLE, ALTER etc.).
DML
Data Manipulation Language – SQL-Befehle zum Einfügen, Ändern und Löschen
von Daten (INSERT, UPDATE, DELETE).
Dump
Ein SQL-Dump ist eine Datei, die den gesamten Inhalt einer Datenbank als SQLBefehle exportiert.
ERD
Entity-Relationship-Diagramm – Grafische Darstellung von Entitäten, Attributen
und Beziehungen in einer Datenbank.
Fremdschlüssel
Foreign Key – Ein Feld, das auf den Primärschlüssel einer anderen Tabelle
verweist und so eine Beziehung herstellt.
GORM
Das Go Object-Relational Mapping – Eine Bibliothek, die Go-Structs auf
Datenbanktabellen abbildet.
Image
Ein Docker-Image ist eine Vorlage für einen Container (z. B. das PostgreSQLImage).
Port
Ein Netzwerk-Port, über den eine Anwendung erreichbar ist (hier: 5432 für
PostgreSQL).
PostgreSQL Ein relationales Open-Source-Datenbank-Management-System (DBMS).
Primärschlüssel Primary Key – Ein eindeutiger Identifikator für einen Datensatz in einer Tabelle.
Begriff Erklärung
Referenzielle
Integrität
Die Korrektheit der Beziehungen zwischen Tabellen – Fremdschlüssel müssen auf
gültige Datensätze verweisen.
Schema
Die Struktur einer Datenbank: alle Tabellen, Spalten, Constraints und
Beziehungen.
Volume
Ein Docker-Volume – ein persistenter Speicherbereich, der auch nach dem
Löschen des Containers erhalten bleibt.