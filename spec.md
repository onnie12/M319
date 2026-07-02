Abschlussauftrag M319 – Der nale Kampf
gegen den Entropie-Drachen
Story-Einstieg
Der eisige Wind des Git-Gletscher-Massivs peitscht den Helden unaufhörlich ins Gesicht. Die
lange, entbehrende Reise nähert sich endlich einem Ende. Die Helden stehen am Fusse des
Vulkanberges der nur The Last Commit genannt wird. In dessen Herz der schwarze Entropie-
Drache seit der schicksalhaften Nacht in der City of All Beginnings sein Nest aufgebaut hatte.
„Sie haben die Ströme geltert, die Blaupausen der Welt gezeichnet, die Logik des Function
Territory gemeistert und das Auge des Debuggers im Wasteland of Exceptions beschworen“,
hallte die Stimme der Entity in ihren Köpfen. „Nun ist es Zeit, das Gelernte zu vereinen. Betreten
Sie den Berg. Stellen Sie sich dem Drachen. Die Runen Ihrer Artefakte sind in der Datenbank
der Welt verewigt – GORM wird Ihre Schöpfungen in die Realität giessen. Möge Ihr Code sauber
sein und Ihre Logik Bestand haben. Das Schicksal von Codera liegt in Ihren Händen – und in
Ihren Goroutines."
1. Überblick
Dieser Gruppenauftrag bildet den Abschluss des Moduls M319. Sie entwickeln ein rundenbasiertes
Kampfsystem in Go, in dem Ihre Gruppe gegen den Entropie-Drachen antritt. Der Drache ist vollständig
von der Lehrperson vorgegeben – Sie implementieren Ihre eigenen Helden-Charaktere, ihre Ausrüstung
und ihre Kampffähigkeiten.
Abgabe: Ein Git-Repository mit vollständigem Go-Projektcode, Godoc-Dokumentation und einer
Gruppendokumentation mit Activity & C4-Diagrammen.
Arbeitsauftrag
Jeder der nachfolgenden Punkte muss nachweisbar erfüllt sein (via Code, Dokumentation oder Git-
History).
Dieser Auftrag läuft parallel zum M164-Auftrag (Datenbanken). Während der M164-Auftrag die
Datenbank-Infrastruktur bereitstellt, wird in diesem Auftrag der Kampf Code implementiert, der die
Logik des Kampfes gegen den Drachen festhält. Die Aufgaben sind so aufgeteilt, dass sie möglichst
gleichzeitig bearbeitet werden können.
Die Rollen fokussieren sich auf unterschiedliche Themen. Nachfolgend eine Auistung für die bessere
Übersicht:
Person M319 (Programmieren) M164 (Datenbanken)
Arkan-
Dokumentar*in C4-Diagramme, Clean Code, Linter Dokumentation des DB-Setups prüfen &
abnehmen
Daten-Druide GORM-Modelle, DB-Connection-
Code DB aufsetzen, Vorgehen dokumentieren
Code-Kleriker*in Logging-Framework DB-Zugangsdaten via .env verteilen
Funktions-
Krieger*in
Kampf-Loop (Integration Helden &
Drache)
DB-Verbindung testen, Query-Ergebnisse
validieren
Runenschmied*in DB-Migration & Seeds im Go-Code SQL-DDL & Seeds parallel als SQL-Skript
vorbereiten
Wichtig: Jedes Gruppenmitglied arbeitet auf seiner eigenen PostgreSQL-Instanz. Jede
Person setzt sich ihren eigenen Container auf (basierend auf der Dokumentation aus dem M164
Auftrag). Der gesamte Go-Code (GORM-Modelle, Seeds, Queries) sowie die SQL-Skripte werden
gegen die persönliche Datenbank ausgeführt. Es gibt keine gemeinsame Datenbank-Instanz,
auf die alle zugreifen.
Teamarbeit & Git
Nutzung von Git (Git Repository) mit einer geeigneten Branching-Strategie (z. B. Git Flow,
GitHub Flow)
Jede/r Lernende ist via Git-History sichtbar an den Arbeiten beteiligt
Commit-Nachrichten sind aussagekräftig und folgen einem einheitlichen Benennungs-Stil
(z.B. Conventional Commits)
⚠ Wichtig – Mitarbeit: Die Git-History muss klar erkennen lassen, wer welche Teile
implementiert hat. Fehlt der Nachweis der Beteiligung einer Person, gibt es einen grossen
Abzug (bis zu einer Note) für die gesamte Gruppe.
Absolutes Minimum: Jede/r Lernende muss seinen/ihren eigenen Charakter (inkl. Stats,
Ausrüstung, Skills und Seed-Daten) vollständig selbst implementieren. Das Commit, das den
eigenen Charakter einführt, muss eindeutig der entsprechenden Person zuordenbar sein. Es
ist nicht erlaubt, dass eine Person den Charakter einer anderen Rolle implementiert – auch
nicht teilweise. Eigenständige Implementierung des eigenen Helden ist die absolute
Mindestanforderung für eine bestandene Mitarbeit.
Jeder Commit muss lauffähig sein: Jeder Commit auf dem Master oder Develop Branch im
Repository muss go build ohne Fehler durchlaufen (auch Zwischenstände). Wenn Code
noch nicht fertig ist, muss er auskommentiert (und mit // TODO bzw. // FIXME markiert)
werden, sodass das Programm kompilierbar bleibt. Einzelne nicht genutzte Variablen/Imports
in Zwischencommits sind tolerierbar, solange der Build nicht bricht. Erst wenn die nale
Abgabe keine Compiler-Fehler mehr enthält, kann sie bewertet werden.
Namenspicht in Seed-Daten: In den Seed-Daten ( db/seeds.go ) muss zwingend der reale
Name der/des Lernenden (oder der Name des Charakters den Sie zu Beginn des Semesters
erstellt haben) als Charaktername verwendet werden (z. B. "Max Mustermann" statt
"Arkan-Dokumentar*in" ). Die Verwendung des Rollennamens statt des echten Namens
führt zu Punktabzug, da die Zuordnung nicht mehr eindeutig nachvollziehbar ist.
Allgemeine Code-Qualität (alle Rollen)
Lauffähiges Go-Programm (go build / go run)
Sinnvolle Paketstruktur mit korrekter Sichtbarkeit (public/private)
Durchgehendes Errorhandling (keine ignorierten Errors)
Saubere Code-Konventionen (Clean Code Regeln werden eingehalten)
Kommentare nur wo nötig (Guter Code braucht fast keine Kommentare!)
Godoc-Dokumentation auf Paketen, Funktionen, Structs, Variablen, Konstanten
Konguration über .env -Files ( .env in .gitignore , .env-example im Repo)
./logs/ -Ordner im .gitignore (Logdateien sollen nicht von Git getrackt werden!)
Unit-Tests für alle zentrale Funktionen (Schadensberechnung, Heilung, Kampf-Logik)
implementiert – go test ./... muss ohne Fehler durchlaufen
Kampf-Logik (Gruppenleistung)
Wichtig: Orientieren Sie sich an den im Moodle beigelegten Beispiel Screenshots wie so ein Kampf
aussehen soll!
Kampf läuft rundenbasiert in der CLI ab
Helden und Drache haben abwechselnd Züge (Reihenfolge basierend auf Initiative/Speed)
RNG-basierte Kampfmechanik (zufällige Schadensstreuung, Ausweichen, kritische Treffer)
Helden können im Kampf sterben (HP ≤ 0)
Kampf endet wenn alle Helden tot sind ODER der Drache besiegt ist
Rollen-spezische Aufgaben
Activity-Diagramm: Jedes Gruppenmitglied erstellt ein Activity-Diagramm seiner eigenen
Rollenimplementation (Aktionsauswahl, Skill-Nutzung, KI-Logik). Die Diagramme werden in der
Gruppendokumentation abgebildet und müssen den Programmablauf der Rolle nachvollziehbar
darstellen.
Rolle Pichtaufgaben Optionale Aufgaben
Arkan-
Dokumentar*in
Charakter-, Ausrüstungs- & Skill-Structs mit Godoc
dokumentieren; C4-Layer 1+2 in der
Gruppendokumentation; Branching-Strategie einrichten
und die Umsetzung kontrollieren; Clean-Code-Regeln
aus dem Cheat Sheet im gesamten Go-Code
durchsetzen; sicherstellen, dass alle Gruppenmitglieder
den Linter installiert haben und der Code bei Abgabe
keine Linting-Errors aufweist
C4-Layer 3
(Component-
Diagramm) als
Bonus; Clean-Code-
Regeln als
.opencode/rule
s.md im Repo
hinterlegen (für den
AI-Assistenten
Opencode)
Code-Kleriker*in
Logging (Debug/Info/Warn/Error) in der Applikation;
Log-Rotation (täglich/wöchentlich); Logle im lokalen
./logs/ -Ordner; Errorhandling + Panic/Recover
Optionale Heil-AI
(Heilung des
schwächsten Helden
sobald jemand unter
30% HP ist)
Runenschmied*in GORM-Modelle für alle Charaktere, Ausrüstung & Skills;
Auto-Migration; Seed-Dateien
Optionale Schutz-AI
(Team-HP-
gesteuerte Schutz-
Rune / Konstrukt-
Schild)
Daten-Druide Datenbank-Seeds mit Bulk-Daten; GORM-Queries für
Ausrüstungs-/Skill-Logik
Optionale Heil-AI
(Selbstheilung unter
40 % HP)
Funktions-
Krieger*in
Goroutines für parallele Kampfaktionen; Mutexe für
Race-Condition-Schutz
Optionale Attack-AI
(Doppelangriff
sobald der Drache
unter 30% HP hat)
System-
Inltrator*in
(Rogue)
Füllt fehlende Rollen in der Gruppe aus;
Implementierung des Rogue-Charakters mit den
stärksten Stats
Optionale
Debuff-/Finish-AI
(Schwachstelle →
Hinterhalt →
Tödliche Präzision)
Abgabetermin
Sehen Sie bei der Abgabe im Moodle & im Moodle Dashboard
2. Vorgegebener Code
Die Lehrperson stellt Ihnen ein Go-Projekt mit folgender Struktur zur Verfügung:
codera-battle/
├── main.go # main()-Funktion (darf angepasst werden)
├── go.mod # Module-Definition
├── .gitignore # Ignoriert .env und logs/
├── .env-example # Für das Spätere Konfigurationsbeispiel (noch
leer)
├── internal/
│ └── types.go # Combatant-Interface + Stats-Struct
├── dragon/
│ └── dragon.go # Vollständige Drachen-Implementierung
└── combat/
└── combat.go # Teilimplementierter Kampf-Loop
main.go
Enthält Platzhalter für .env -Loading
Enthält Platzhalter für Logging-Initialisierung
Enthält Platzhalter für Datenbankverbindung, Auto-Migration & Seeds
Enthält Platzhalter zum Laden der Helden aus der DB
Erstellt den Drachen (aus Konstanten, keine DB nötig)
Startet den Kampf ( CombatLoop )
Sie dürfen main.go anpassen/ergänzen, solange die Grundstruktur erhalten bleibt
internal/types.go
Stellt das Combatant -Interface und das Stats -Struct bereit, die sowohl von Helden als auch vom
Drachen implementiert werden:
type Combatant interface {
GetName() string
GetStats() Stats
GetCurrentHP() int
SetCurrentHP(hp int)
GetMaxHP() int
IsAlive() bool
}
dragon/dragon.go (vollständig – nicht verändern)
Dieser Absatz dient nur der Dokumentation (für Ihr Verständnis). Es soll Ihnen primär als Beispiel dienen
(die Implementation des Drachens und der Helden ist sich sehr ähnlich).
Der Drache wird in lokalen/globalen Variablen und Konstanten gespeichert (keine Datenbank nötig). Er
ist vollständig implementiert inklusive:
Stats:
HP: 450
Attack: 30
Defense: 18
Speed: 14
Fähigkeiten:
Skill Beschreibung Schaden Genauigkeit Typ
Entropy Claw Krallenangriff 18–32 90 % Einzelziel
Null Pointer Breath Entropie-Atem 24–42 75 % Einzelziel
Stack Overow Flächenangriff (alle Helden) 12–22 60 % AoE
Corrupted Code Drache heilt sich – (heilt 20) 100 % Selbstheilung
Rage (passiv) Ab 30 % HP aktiv: +50 % Schaden – – passiv
KI-Verhalten:
Normalmodus: zufällige Skill-Auswahl
Rage-Modus: bevorzugt offensive Skills
Notheilung: unter 20 % HP → 50 % Chance auf Heilung
Corrupted Code wird nur alle 4 Runden eingesetzt
Thread-Safety: Alle HP-Änderungen sind via sync.Mutex geschützt
(als Vorbereitung für optionale Goroutines des Funktions-Kriegers).
combat/combat.go (teilimplementiert)
Vorgegeben:
CombatLoop(heroes, dragon) – Hauptkampf-Loop mit Rundenverwaltung und
Sieg/Niederlage-Prüfung
buildInitiativeOrder() – Sortiert alle Teilnehmer nach Speed (absteigend)
CalculateDamage() – Schadensformel inkl. RNG, Genauigkeit, Kritischen Treffern (darf nicht
verändert werden)
processDragonTurn() – Basis-KI für den Drachen (muss erweitert werden)
critSuffix() , printBattleResult() – Hilfsfunktionen
Von Lernenden zu implementieren:
processHeroTurn() – Komplette Neuimplementierung: CLI-Anzeige, Aktionsauswahl, Skill-
Nutzung
processDragonTurn() – Erweiterung: Korrekte Anwendung von ChooseAction() , Rage-
Bonus, Mutexe
logAction() – Logging aller Kampfaktionen
3. Charakter-Design: Rollen und ihre Specs
Jede/r Lernende implementiert ihren/seinen Helden-Charakter in einem separaten Paket
( hero/<rolle>/ ). Jeder Charakter besteht aus:
3.1 Gemeinsame Basis-Structs
Vorgegeben in internal/types.go (durch die Lehrperson):
Stats – enthält die Kampfwerte MaxHP, Attack, Defense, Speed.
Combatant – Interface mit GetName(), GetStats(), GetCurrentHP(), SetCurrentHP(), GetMaxHP(),
IsAlive().
Von den Lernenden zu denieren – leiten Sie die Structs aus den Anforderungen der Abschnitte 3.2–3.7
ab:
Equipment – Ein Gegenstand mit Namen, Typ (weapon/armor/accessory), Stat-Boni und
optionalem Spezialeffekt. Überlegen Sie, wie Sie die Boni abbilden (einzelne Attribute oder ein
eingebetteter Stats-Struct).
Skill – Eine Fähigkeit mit Namen, Schadensbereich (Min/Max), Heilungswert, Genauigkeit (0.0–1.0),
Zieltyp (single_enemy, all_enemies, single_ally, all_allies, self) und Beschreibung.
Diese Structs dienen als Grundlage für die Helden-Implementierung in den folgenden Abschnitten und
unterscheiden sich von den GORM-Datenbankmodellen aus Abschnitt 5.1.
3.2 Rollen-spezische Vorgaben
Jede/r Lernende implementiert GENAU die Struktur und Skills, die zu ihrer/seiner Rolle passen.
Arkan-Dokumentar*in (Magier*in) – «Runen der Offenbarung»
Basis-Stats:
HP: 120
Attack: 18
Defense: 8
Speed: 14
Ausrüstung:
Gegenstand Typ Stat-Bonus Spezialeffekt
Pergament-Stab weapon +8 Attack –
Runen-Gewand armor +5 Defense –
Tintenfass-Amulett accessory +3 Speed, +20 MaxHP –
Skills:
Skill Schaden Heilung Genauigkeit Ziel Beschreibung
Runen-Geschoss 12–24 0 90 % single_enemy Magischer
Runenangriff
Arkaner Bann 8–16 0 85 % all_enemies Schwacher
Flächenangriff
Klärende Annotation
(Heilung) 0 15–25 100 % single_ally Heilt einen
Verbündeten
Optional: Kann im Kampf basierend auf dem niedrigsten HP-Wert im Team automatisch den Heilzauber
wählen (sonst offensiv). Dokumentieren Sie in Godoc die gewählte Strategie.
Daten-Druide (Formwandler*in) – «Ströme der Transformation»
Basis-Stats:
HP: 100
Attack: 14
Defense: 10
Speed: 16
Ausrüstung:
Gegenstand Typ Stat-Bonus Spezialeffekt
Transformations-Kristall weapon +6 Attack –
Datenstrom-Mantel armor +4 Defense –
Gegenstand Typ Stat-Bonus Spezialeffekt
Schema-Ring accessory +5 Speed, +10 MaxHP –
Skills:
Skill Schaden Heilung Genauigkeit Ziel Beschreibung
Datenklinge 10–20 0 85 % single_enemy Transformierte Daten
als Klinge
Strukturwandel 14–28 0 70 % single_enemy Hoher Schaden,
niedrige Genauigkeit
Transformative
Regeneration 0 12–20 100 % self Heilt sich selbst
Optional: Wechselt die Skill-Wahl basierend auf dem eigenen HP-Level (unter 40 % Heilung, sonst
offensiv).
Code-Kleriker*in (Heiler*in) – «Licht des Debuggers»
Basis-Stats:
HP: 110
Attack: 10
Defense: 12
Speed: 12
Ausrüstung:
Gegenstand Typ Stat-Bonus Spezialeffekt
Debugger-Stab weapon +4 Attack –
Kleriker-Robe armor +6 Defense –
Auge-des-Debuggers-Amulett accessory +2 Speed, +30 MaxHP –
Skills:
Skill Schaden Heilung Genauigkeit Ziel Beschreibung
Heiliges Licht 6–12 0 95 % single_enemy Schwacher Lichtangriff
Heilsame
Korrektur 0 20–35 100 % single_ally Starke Heilung eines
Verbündeten
Skill Schaden Heilung Genauigkeit Ziel Beschreibung
Segen der
Stabilität 0 10–15 100 % all_allies Flächenheilung (alle
Helden)
Zusätzliche Aufgaben:
Hier muss zwingend das .env -basierte Logging eingebaut sein.
Beim Programmstart muss geprüft werden, ob das Logle beschreibbar ist. Falls nicht: panic
mit aussagekräftiger Fehlermeldung.
Beim Programmstart muss geprüft werden, ob die Datenbankverbindung hergestellt werden kann.
Falls nicht: panic mit aussagekräftiger Fehlermeldung.
Optional: Überprüft vor jeder Aktion die HP aller Helden. Wenn ein Verbündeter unter 30 % HP ist, wird
automatisch Heilsame Korrektur auf ihn gewirkt, anstatt anzugreifen. Wenn mehrere Helden
unter 30 % sind, wird zuerst der mit den wenigsten HP geheilt. Sind alle über 30 %: Heilung auf den
Helden mit den wenigsten HP ODER Angriff (50:50).
Funktions-Krieger*in (Warrior) – «Präzision der Funktionen»
Basis-Stats:
HP: 150
Attack: 22
Defense: 14
Speed: 8
Ausrüstung:
Gegenstand Typ Stat-Bonus Spezialeffekt
Funktions-Schwert weapon +10 Attack –
Krieger-Rüstung armor +8 Defense –
Gurt der Ausdauer accessory +2 Speed, +40 MaxHP –
Skills:
Skill Schaden Heilung Genauigkeit Ziel Beschreibung
Präziser
Hieb 18–32 0 80 % single_enemy Kräftiger physischer Angriff
Schutzschild 0 0 100 % self Erhöht eigene Defense um 5 für
diese Runde
Skill Schaden Heilung Genauigkeit Ziel Beschreibung
Kampfschrei 8–16 0 90 % single_enemy Schwächerer Angriff, aber +5
Attack nächste Runde
Optional: Der Krieger führt einen Double Strike aus – zwei Präziser Hieb -Angriffe gleichzeitig über
Goroutines. Ein sync.WaitGroup wartet, bis beide Schadensberechnungen abgeschlossen sind,
bevor der Schaden auf den Drachen angewendet wird. Ein Mutex schützt die gemeinsame HP-
Verwaltung des Drachens. Zusätzlich: Wählt Schutzschild automatisch wenn die eigene HP unter
30 % fällt.
Runenschmied*in (Klassen-Architekt) – «Architektur der Macht»
Basis-Stats:
HP: 130
Attack: 16
Defense: 16
Speed: 10
Ausrüstung:
Gegenstand Typ Stat-Bonus Spezialeffekt
Architekten-Hammer weapon +7 Attack –
Runen-Plattenpanzer armor +9 Defense –
Siegelring der Stabilität accessory +1 Speed, +25 MaxHP –
Skills:
Skill Schaden Heilung Genauigkeit Ziel Beschreibung
Architekten-
Schlag 14–26 0 85 % single_enemy Solider physischer Angriff
Schutz-Rune 0 0 100 % all_allies Erhöht Defense aller Helden
um 3 für eine Runde
Konstrukt-
Schild 0 0 100 % single_ally Ein Verbündeter erhält -50 %
Schaden für 1 Runde
Optional: Setzt Schutz-Rune automatisch ein, wenn das durchschnittliche Team-HP unter 50 % fällt.
Der schwächste Verbündete erhält Konstrukt-Schild , wenn er unter 25 % HP ist.
System-Inltrator*in (Rogue) – Stärkste Klasse (nur für besondere Gruppen)
Hinweis: Der Rogue hat alle Zusatzaufträge gemeistert und ist kämpferisch die stärkste
Klasse. In der Gruppe füllt er jene Rollen, die von niemand anderem abgedeckt werden. Seine
hohen Werte gleichen das fehlende Teammitglied aus.
Basis-Stats (mit Bonus):
HP: 120
Attack: 30
Defense: 10
Speed: 20
Ausrüstung:
Gegenstand Typ Stat-Bonus Spezialeffekt
Schatten-Dolch weapon +14 Attack life_steal (10 % des Schadens als
Heilung)
Inltrator-Cape armor +5 Defense –
Amulett der
Verwundbarkeit accessory +5 Speed, +25
MaxHP –
Skills:
Skill Schaden Heilung Genauigkeit Ziel Beschreibung
Hinterhalt 22–40 0 80 % single_enemy Hoher Schaden, mittlere
Genauigkeit
Schwachstelle
analysieren 0 0 100 % single_enemy
Senkt die Defense des
Drachens um 5 für 2
Runden
Tödliche Präzision 18–34 0 90 % single_enemy Wenn der Drache unter 25
% HP: doppelter Schaden
Optional: Setzt zuerst Schwachstelle analysieren ein (solange der Drache keinen Debuff hat),
danach Hinterhalt . Wenn der Drache unter 25 % HP fällt, wird nur noch Tödliche Präzision
verwendet.
4. Kampfsystem – Detaillierte Spezikation
4.1 Ablauf einer Kampfrunde
. Initiative bestimmen: Alle Teilnehmer (Helden + Drache) werden nach Speed (absteigend) sortiert.
Bei Gleichstand mit dem Drachen: Helden zuerst. (Bei mehreren Helden zufällig)
. Zug ausführen: In der Reihenfolge führt jeder Teilnehmer eine Aktion aus.
. Rundenende: Prüfen, ob Kampf beendet ist.
. Nächste Runde: Wiederholen, bis Kampf endet.
4.2 Helden-Zug (CLI-Interaktion)
╔══════════════════════════════════════════╗
║ Entropie-Drache - HP: 187/300 ║
║ Status: Rage aktiv (+50% Schaden) ║
╚══════════════════════════════════════════╝
Runde 3 - Zug von Arkan-Dokumentar*in (Magier*in)
─── Team HP ───
Magier*in: 85/120 ▼
Druide: 64/100
Kleriker*in: 103/110 ▼▼
Krieger*in: 112/150 ▼
─── Aktionen ───
1. Runen-Geschoss (12-24 Schaden, 90%)
2. Arkaner Bann (8-16, alle Gegner, 85%)
3. Klärende Annotation (heilt 15-25)
Deine Wahl (1-3):
Hinweis: Wenn einer der Automatismen ausgeführt werden soll (z.B. Automatische Heilung des
Klerikers) dann bekommt der entsprechende Held keine Auswahl sondern der Zug wird für ihn
ausgeführt (und das dann in der CLI angezeigt).
4.3 Schadensberechnung
func CalculateDamage(baseMin, baseMax int, attackerStat, defenderDef int,
// 1. Genauigkeits-Check (RNG)
if rand.Float64() > accuracy {
return 0, false, true // Angriff verfehlt
}
// 2. Basisschaden (RNG innerhalb der Spanne)
baseDamage := rand.Intn(baseMax-baseMin+1) + baseMin
4.4 Drachen-KI (vorgegeben in dragon.go )
Normalmodus (HP > 30 %): zufällige Skill-Auswahl (alle Skills gleich wahrscheinlich, aber
Corrupted Code nur alle 4 Runden)
Rage-Modus (HP ≤ 30 %): Enraged = true , +50 % Skill-Schaden. Der Drache versucht zu
heilen wenn HP < 20 % (50 % Chance statt anderem Angriff)
5. Datenbank (GORM) – Vorgaben
Hinweis: In den vorgegebenen Code-Dateien ist keine Datenbankverbindung enthalten. Die
Datenbank (GORM + PostgreSQL) muss von Ihnen selbst aufgebaut werden. Nur die Charakter-
Attribute der Lernenden (Helden, Ausrüstung, Skills) werden in der Datenbank gespeichert.
Der Drache ist in Konstanten deniert ( dragon.go ) und benötigt keine Datenbank.
5.1 Modelle
// 3. Angriffsbonus (Attacker-Stat / 10 als Multiplikator)
attackMultiplier := 1.0 + float64(attackerStat)/20.0
// 4. Verteidigungsreduktion
defenseReduction := 1.0 - float64(defenderDef)/100.0
if defenseReduction < 0.1 {
defenseReduction = 0.1 // Minimum 10% Schaden kommen immer durch
}
finalDamage := int(float64(baseDamage) * attackMultiplier * defenseRe
if finalDamage < 1 {
finalDamage = 1 // Minimum 1 Schaden
}
// 5. Kritischer Treffer (10% Chance, 1.5x Schaden)
isCrit := rand.Float64() < 0.1
if isCrit {
finalDamage = int(float64(finalDamage) * 1.5)
}
return finalDamage, isCrit, false
}
Leiten Sie aus den Rollen-Beschreibungen in Abschnitt 3 und den Tabellen aus dem M164-Auftrag Ihre
GORM-Modelle selbst ab. Sie benötigen drei Modelle – überlegen Sie, welche Entitäten und Attribute
nötig sind:
Hero: Ein Held mit Namen, Rolle, Kampfwerten (MaxHP, CurrentHP, Attack, Defense, Speed) sowie
Referenzen auf drei Ausrüstungsgegenstände (Waffe, Rüstung, Accessoire).
Equipment: Ein Ausrüstungsgegenstand mit Name, Typ, Boni (Attack, Defense, Speed, HP) und
optionalem Spezialeffekt.
Skill: Eine Fähigkeit mit Name, zugehöriger Rolle, Schadensbereich, Heilungswert, Genauigkeit,
Zieltyp und Beschreibung.
Hinweise:
Betten Sie gorm.Model ein (damit ID, CreatedAt, UpdatedAt, DeletedAt automatisch erzeugt
werden).
Verwenden Sie GORM-Tags ( gorm:"foreignKey:..." ), um die Beziehungen zwischen Hero
und Equipment abzubilden.
Ein Held kann maximal drei Gegenstände tragen – überlegen Sie, ob Sie einzelne Fremdschlüssel
oder eine viele-zu-viele-Beziehung (Zwischentabelle) verwenden möchten.
5.2 Seeds
Die Seed-Datei muss alle Helden, Ausrüstungsgegenstände und Skills jedes Gruppenmitglieds in die
Datenbank einfügen. Die Seeds werden beim Start von main.go ausgeführt. Die Daten müssen mit
den Vorgaben aus Abschnitt 3 übereinstimmen.
6. Zusatzthemen-Matrix
Rolle Zusatzthema Wo im Auftrag?
Arkan-
Dokumentar*in
Git &
Branching-
Strategie
Git-History, Branches
Arkan-
Dokumentar*in
Paketstruktur
&
Sichtbarkeiten
Aufteilung in internal/ , pkg/ etc.
Arkan-
Dokumentar*in Godoc Alle exportierten Symbole dokumentiert
Arkan-
Dokumentar*in C4-Modell Gruppendokumentation
Rolle Zusatzthema Wo im Auftrag?
Code-Kleriker*in Logging (4
Stufen)
log.Debug() , log.Info() , log.Warn() ,
log.Error() im Kampfablauf
Code-Kleriker*in Log-Rotation
Log-Datei im ./logs/ -Ordner wird täglich rotiert (z. B. via
lumberjack oder eigenem Rotator); ./logs/ im
.gitignore
Code-Kleriker*in .env-
Konguration Log-Level, Log-Pfad (auf ./logs/battle.log ) via .env
Code-Kleriker*in
Errorhandling
+
Panic/Recover
Graceful handling bei DB-Fehlern, Panic/Recover in Goroutines
Runenschmied*in GORM Hero/Equipment/Skill als GORM-Modelle
Runenschmied*in Auto-
Migration
db.AutoMigrate(&Hero{}, &Equipment{},
&Skill{})
Runenschmied*in Seeds Seed-Files mit Initialdaten
Daten-Druide GORM-Queries Datenbankabfragen für Ausrüstungs-/Skill-Zuweisung
Daten-Druide Seeds (Bulk-
Daten) Bulk-Insert aller Items/Skills
Funktions-
Krieger*in Goroutines Parallele Schadensberechnung, parallele Heilung
Funktions-
Krieger*in Mutexe Schutz der Dragon-HP bei parallelen Zugriffen
7. Lieferumfang (Checkliste)
Wichtig: Alle geforderten Elemente sind als Zip File im Moodle abzugeben. Die
Gruppendokumentation (mit dem Link zum Repository) geben Sie aber separat ab.
Git-Repository mit vollständiger History und Branching-Strategie (Download des Repos als Zip
File + Link zum Repo)
main.go startet das Programm (inkl. nötiger Anpassungen)
dragon.go (unverändert von Lehrperson)
combat/combat.go (Kampf-Loop komplett implementiert)
hero/<rolle>/ (eigenes Paket pro Rolle (1 pro Lernenden in der Gruppe))
db/models.go (GORM-Modelle)
db/seeds.go (Seed-Daten mit echten Namen (oder Charakternamen) der
Gruppenmitglieder)
.env-example (Kongurationsvorlage mit Beispielwerten)
.gitignore (mit .env und logs/ )
Unit-Tests ( *_test.go ) für Schadensberechnung, Heilung und Kampf-Logik
go.mod + go.sum (werden automatisch generiert, müssen aber mitabgegeben werden)
README.md mit Kurzbeschreibung des Programms (für eine grundlegende Übersicht)
Gruppendokumentation (inkl. Activity- und C4-Diagrammen)
Godoc-Dokumentation ( godoc -http :8080 muss funktionieren)
Glossar
Begriff Erklärung
AoE Area of Effect – Ein Angriff oder Effekt, der alle Gegner (oder alle Verbündeten)
gleichzeitig trifft, nicht nur ein einzelnes Ziel.
CLI Command-Line Interface – Die Bedienung des Programms erfolgt über die
Kommandozeile (Terminal) durch Texteingabe.
C4-Modell Eine Methode zur Architektur-Darstellung in 4 Abstraktionsebenen: Context,
Container, Component, Code.
Concurrency Nebenläugkeit – Mehrere Abläufe werden zeitlich verschränkt ausgeführt. In Go
umgesetzt mit Goroutines.
DML / DDL Data Manipulation Language (INSERT, UPDATE, DELETE) vs. Data Denition
Language (CREATE TABLE, ALTER, etc.).
ERD Entity-Relationship-Diagramm – Grasche Darstellung von Entitäten, Attributen und
Beziehungen in einer Datenbank.
GORM Das Go Object-Relational Mapping – Eine Bibliothek, die Go-Structs auf
Datenbanktabellen abbildet.
Godoc Go-Dokumentations-Tool – Erzeugt automatisch Dokumentation aus Code-
Kommentaren.
Goroutine Ein leichtgewichtiger, nebenläuger Ausführungsfaden in Go. Wird mit go func()
gestartet.
Initiative Die Reihenfolge, in der Kämpfer in einer Runde handeln. Bestimmt durch den Speed-
Wert (höher = früher dran).
Linter Ein Werkzeug, das Code auf stilistische und syntaktische Fehler prüft (z. B.
golangci-lint ).
Mutex Mutual Exclusion – Ein Synchronisationsmechanismus, der verhindert, dass mehrere
Goroutines gleichzeitig auf dieselbe Variable zugreifen.
RNG Random Number Generator – Zufallszahlengenerator, z. B. für Schadensstreuung,
Genauigkeit und kritische Treffer.
Begriff Erklärung
Seed-Daten Initiale Daten, mit denen die Datenbank beim Start befüllt wird (Helden, Ausrüstung,
Skills).
WaitGroup Ein Synchronisationswerkzeug in Go, das darauf wartet, dass eine Sammlung von
Goroutines beendet ist.