# Codera Battle – Vorgegebener Code

Dieses Paket enthält das Grundgerüst für den Abschlussauftrag M319.

Passen Sie dieses README an sobald Sie mit dem Projekt fertig sind. Es soll einem Programmierenden einen grundlegenden Überblick über die Funktion und die Struktur des Programms geben!

## Verzeichnisstruktur

```
.
├── main.go            # Einstiegspunkt (darf angepasst werden)
├── go.mod             # Module-Definition
├── go.sum
├── internal/
│   └── types.go       # Combatant-Interface & Stats (nicht verändern)
├── dragon/
│   └── dragon.go      # Entropie-Drache (vollständig, nicht verändern)
└── combat/
    └── combat.go      # Kampf-Loop (teilimplementiert)
```

## Was ist bereits fertig?

- **Drache** mit KI, Fähigkeiten und Rage-Modus (Konstanten, keine DB nötig)
- **Schadensformel** mit RNG, Genauigkeit und kritischen Treffern
- **Kampf-Loop** mit Initiativereihenfolge und Sieg/Niederlage-Prüfung
- **Platzhalter-Helden** damit das Programm startet
