# 🥊 Street FACS Fighter & Terminal FACS Atlas

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-black?logo=linux)](https://github.com)
[![Protocol](https://img.shields.io/badge/Terminal-Sixel%20Graphics-orange)](https://github.com/mattn/go-sixel)

> **"Can you decode human micro-expressions before time runs out?"**  
> An arcade-style, terminal-native combat game and interactive diagnostic atlas based on Paul Ekman & Wallace V. Friesen's **Facial Action Coding System (FACS)**. Powered by high-resolution inline Sixel graphics, chiptune sound effects, and embedded master anatomical rationale datasets.

---

## 🌟 Overview

**Street FACS Fighter** reimagines facial expression analysis as a high-intensity terminal arcade experience and a modern CLI alternative to legacy FACS viewers. 

Faces are rendered directly in your terminal using **Sixel graphics**. Players must observe subtle muscle contractions, identify the corresponding **Action Units (AUs)**, and execute keystrokes under real-time Active Time Battle (ATB) pressure before the opponent unleashes a devastating Rage Attack.

The program also functions as **Street FACS Atlas / Manual**—a full-featured, zero-dependency interactive viewer to inspect, search, and study official master scoring examples alongside Paul Ekman's anatomical rationales.

---

## ✨ Key Features

- **🖼️ Native In-Terminal Sixel Rendering**
  - No legacy Java applets, browser plugins, or external GUI windows required.
  - High-definition portrait rendering scaled dynamically to fit standard terminal windows without vertical scroll overflow.
  - Reactive visual effects: red tint flash on enemy rage detonation and screen impacts.
- **🥋 Strict FACS Official Input Validation (Hardcore Mode)**
  - Bare numbers (e.g., `12` or `1 4`) are rejected as non-standard syntax.
  - Enforces official FACS prefixes: Action Units (`AU12`, `AU4`), Action Descriptors (`AD19`, `AD38`), Head Movements (`M55`, `M68`), and lateralized unilateral expressions (`LAU12`, `RAU14`, `L12`, `R14`).
  - Supports rapid multi-AU combo inputs separated by spaces or commas (e.g., `AU1 AU4` or `LAU12, AU25`).
- **🧠 Embedded FACS Master Dataset (`//go:embed`)**
  - Built-in TSV dataset containing scores, media paths, item IDs, splits, and comprehensive anatomical rationales.
  - Standalone SQLite3 ingestion without external file dependencies.
- **💡 Practice & Training Mode (`-training` / `practice`)**
  - Pauses enemy ATB attacks entirely.
  - Displays correct target Action Units, muscle descriptions, and concise Ekman guidance at the top of the HUD for stress-free observational training.
- **📖 Street FACS Manual / Interactive Atlas Mode (`manual`)**
  - Browse all reference expressions interactively with `[N]`ext, `[P]`rev, and `[Q]`uit.
  - Filter specific Action Units on demand (e.g., `-manual -au 12`).
  - Read full, unclipped anatomical explanations explaining why particular facial furrows, lid tighteners, or lip positions were scored.
- **🎵 Synthesized Chiptune Audio & BGM Engine**
  - Bundled mathematical waveform synthesizer (`cmd/gen-se`) generates 8-bit PCM WAV sound effects (`hit.wav`, `damage.wav`, `ko.wav`, `lose.wav`) on the fly.
  - Seamless background music playback through standard CLI audio backends (`mpv`, `ffplay`, `pw-play`, `paplay`, `aplay`).
- **📊 Post-Battle Diagnostic Report & Analyst Ranking**
  - After battle or defeat, review each stage with Sixel thumbnails, official FACS scores, hit/miss logs, and Ekman's full anatomical rationales.
  - Performance-based evaluation system awarding certified ranks from **Rank C (Novice)** up to **Rank S (Certified FACS Master)**.
- **🌐 Full Bilingual Internationalization (i18n)**
  - Seamlessly switch between Japanese and English from the title screen or CLI flag (`-lang en` / `-lang ja`).

---

## 💻 Requirements

1. **Go**: Version 1.22 or newer
2. **Sixel-Compatible Terminal Emulator**:
   - `foot`, `wezterm`, `kitty` (with sixel support), `alacritty` (with sixel patch), `mintty`, `xterm`
3. **Audio Player (Optional, for BGM / Sound Effects)**:
   - Any of `mpv`, `ffplay`, `pw-play` (PipeWire), `paplay` (PulseAudio), or `aplay` (ALSA)

---

## 🚀 Quick Start

### 1. Synthesize 8-Bit Sound Effects

Generate the retro sound effects (Hit, Damage, KO, and Defeat) using the built-in mathematical synthesizer:

```bash
go run ./cmd/gen-se
```
*(Writes `hit.wav`, `damage.wav`, `ko.wav`, and `lose.wav` to `assets/sounds/`)*

### 2. Ingest Embedded FACS Master Dataset

Synchronize the embedded master dataset into the local SQLite database (`app.db`):

```bash
go run ./cmd/sff-importer
```

### 3. Launch the Game

```bash
go run ./cmd/street-facs-fighter
```

---

## 🎮 Launch Modes & Usage

### 🥊 Arcade Battle Mode
Face a randomized series of expressions. Strike target Action Units before the opponent's ATB gauge reaches 100%:

```bash
# Default (Japanese UI)
go run ./cmd/street-facs-fighter

# English UI
go run ./cmd/street-facs-fighter -lang en
```

### 💡 Practice & Training Mode
Train without enemy attacks. Correct AUs, muscle names, and anatomical hints remain visible on-screen:

```bash
go run ./cmd/street-facs-fighter -training
# or
go run ./cmd/street-facs-fighter practice
```

### 📖 Street FACS Manual / Atlas Browser
Inspect the complete dataset as an interactive visual encyclopedia:

```bash
# Browse all items
go run ./cmd/street-facs-fighter -manual
# or
go run ./cmd/street-facs-fighter manual

# Filter by a specific Action Unit (e.g., AU12 - Zygomaticus major)
go run ./cmd/street-facs-fighter -manual -au 12
```

---

## 🥋 Strike Input Syntax Rules

To mirror official FACS scoring rigor, bare numbers are marked as syntax errors. Use valid prefixes:

| Input Example | Status | Description |
| :--- | :---: | :--- |
| `AU12` | ⭕ Valid | Action Unit 12 (*Zygomaticus major* / Lip Corner Puller) |
| `AU1 AU4` | ⭕ Valid | Simultaneous combo targeting AU1 and AU4 |
| `AU6, AU12, AU25` | ⭕ Valid | Comma-delimited multi-AU combo |
| `LAU12` / `L12` | ⭕ Valid | Unilateral left AU12 |
| `RAU14` / `R14` | ⭕ Valid | Unilateral right AU14 (*Buccinator* / Dimpler) |
| `AD19` / `M55` | ⭕ Valid | Action Descriptors / Gross Head Movements |
| `12` / `1 4` | ❌ Invalid | Missing prefix; counts as a miss and spikes enemy ATB |
| `?` or `HINT` | 💡 Reference | Displays the Action Unit cheat sheet matrix |

---

## 📂 Project Structure

```text
.
├── cmd/
│   ├── gen-se/              # 8-bit algorithmic WAV sound synthesizer
│   ├── sff-importer/        # Embedded TSV master importer to SQLite3
│   └── street-facs-fighter/ # Game engine & interactive Atlas CLI
├── internal/
│   ├── audio/               # BGM & SE playback controller
│   ├── domain/
│   │   ├── entity/          # Question & TargetAU domain entities
│   │   ├── locale/          # Bilingual localization (JA / EN) & AU dictionaries
│   │   └── repository/      # Repository & transaction interfaces
│   ├── infrastructure/
│   │   └── sqlite3/         # SQLite3 persistence & transaction manager
│   ├── ui/
│   │   ├── input/           # Raw terminal mode & non-blocking input listener
│   │   └── terminal/        # Inline Sixel rendering, rage shaders & text wrapping
│   └── usecase/             # Combat loop, manual viewer, and import usecases
└── assets/
    ├── sounds/              # BGM (MP3) and synthesized effects (WAV)
    └── *.gif                # Practice facial expressions & animations
```

---

## ⚠️ Disclaimer

- This software is developed solely for educational, psychological, and cognitive neuroscience training purposes.
- This repository does not redistribute proprietary examination materials or copyrighted reference manuals. Users may supply authorized local training images under `assets/`.

---

## 📜 License

MIT License. See [LICENSE](LICENSE) for details.