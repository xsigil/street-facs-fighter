# 🥊 Street FACS Fighter

> **The Ultimate Action Unit Battle Arena in Your Terminal**  
> Master facial muscle decoding (FACS) through real-time ATB combat and high-resolution Sixel graphics in a retro arcade-style terminal fighting game!

---

## 🌟 Overview

**Street FACS Fighter** is a terminal-based arcade combat game themed around the **Facial Action Coding System (FACS)**—the anatomical and psychological standard for categorizing human facial movements.

Opponents' facial expressions are rendered directly inside your terminal emulator with high-fidelity **Sixel graphics**. Players must decode active facial muscles (**Action Units / AUs**) in real time and execute keystroke strikes before the enemy's Active Time Battle (ATB) gauge fills. When an opponent's facial tension reaches 100%, they unleash a devastating **Rage Attack**.

---

## ✨ Features

- **🖼️ Native In-Terminal Sixel Image Rendering**
  - High-resolution rendering directly inside modern terminal emulators.
  - Dynamic image filters: opponent portraits pulse with tension as ATB builds, flash with red shadows, and shake during impact.
- **⚡ Real-Time Active Time Battle (ATB) Engine**
  - High-intensity pressure mechanics where enemy gauge ticks up dynamically.
  - Multi-target combo strikes supported via space- or comma-separated tokens (e.g., entering `1 4` or `12 25`).
- **🎵 Seamless Audio & Sound Effects Engine**
  - Automatic background playback integration with popular media players (`mpv`, `ffplay`, `pw-play`, `paplay`).
  - Bundled mathematical waveform synthesizer (`cmd/gen-se`) generates punchy retro hit sounds, heavy damage impacts, victory fanfares, and defeat jingles on demand.
- **📊 Post-Battle Review & Diagnostic Report**
  - Displays Sixel thumbnails of each stage's face, correct AU targets, hit and miss history, strike accuracy percentage, and overall score.
  - Awards official **FACS Analyst Ranks** (from Rank C apprentice to Rank S certified master).
- **🌐 Full Bilingual Internationalization (i18n)**
  - Toggle between English and Japanese on the fly from the title screen (`E` / `J`) or via CLI flag (`-lang en` / `-lang ja`).
  - Anatomical muscle names (*Frontalis, pars medialis*, *Zygomaticus major*, etc.) fully translated in both languages.

---

## 💻 Requirements

1. **Go**: Version 1.20 or newer
2. **Sixel-Compatible Terminal Emulator**:
   - [WezTerm](https://wezfurlong.org/wezterm/) (Recommended)
   - [iTerm2](https://iterm2.com/)
   - [foot](https://codeberg.org/dnkl/foot)
   - [mlterm](https://mlterm.sourceforge.net/)
   - Windows Terminal (Preview with Sixel enabled)
3. **Audio Player (Optional)**:
   - Any one of `mpv`, `ffplay` (from ffmpeg), `pw-play` (PipeWire), or `paplay` (PulseAudio)

---

## 🚀 Quick Start

### 1. Clone Repository & Install Dependencies

```bash
git clone https://github.com/<YOUR_USER>/street-facs-fighter.git
cd street-facs-fighter
go mod tidy
```

### 2. Synthesize Sound Effects

Generate retro 8-bit sound effects instantly using the built-in mathematical synthesizer:

```bash
go run ./cmd/gen-se
```
*(Creates `hit.wav`, `damage.wav`, `ko.wav`, and `lose.wav` under `assets/sounds/`)*

### 3. Place FACS Practice Images & Import to Database

Place your FACS training practice image files (e.g., from the manual practice sets or digital archives) directly into `assets/`, then build the SQLite index:

```bash
# Ingest and parse all expression images into app.db
make import
```
*(Alternatively, execute `go run ./cmd/sff-importer` directly)*

### 4. Configure BGM Tracks (Optional)

Place your stage music (`stage1.mp3`) and results music (`lose.mp3`) under `assets/sounds/`:

```bash
# Automated deployment helper for local or downloaded tracks
chmod +x scripts/setup_audio.sh
./scripts/setup_audio.sh
```

### 5. Launch the Game

```bash
# Launch in English mode directly
go run ./cmd/street-facs-fighter -lang en

# Or launch with default settings
make run
```

---

## 📸 Dataset Import & Filename Conventions

The game includes an automated expression importer (`cmd/sff-importer`) that scans `assets/` and indexes images into `app.db`. It automatically parses the standardized file naming conventions used across FACS training datasets and digital practice archives:

```text
assets/
├── s1.gif             -> AU1 (Inner Brow Raiser)
├── s4a.gif            -> AU4 (Brow Lowerer, variant a)
├── s1_4a.gif          -> AU1 + AU4 (Multi-AU combo target)
├── s6_12y25.gif       -> AU6 + AU12 + AU25 (Smile with cheek raise and parted lips)
├── s10y_15z.gif       -> AU10 + AU15
├── sW4_5x.gif         -> AU4 + AU5 (Model set W)
└── sL20x_26.gif       -> AU20 + AU26 (Model set L)
```

### Parsing Rules

| Filename Component | Meaning | Example |
| :--- | :--- | :--- |
| **Prefix (`s`, `sW`, `sL`, `sJ`)** | Subject / sample series indicator | `s`, `sW`, `sL` are stripped during parsing |
| **Delimiters (`_`)** | Separates discrete AU combinations | `s1_4a.gif` targets AU `1` and AU `4` |
| **Fused Identifiers** | Multi-digit combinations without underscores | `s1012x25.gif` extracts AU `10`, `12`, and `25` |
| **Intensity & Variants (`a`, `b`, `x`, `y`, `z`)** | FACS intensity scores / variant markers | Stripped automatically to identify the base AU code |

Once indexed into SQLite (`app.db`), the battle engine randomly selects targets per stage and verifies keystrokes against all extracted Action Units.

---

## 🎮 How to Play

- **Strike Weak-Point AUs**:
  - `12` + `Enter`: Single strike targeting AU12 (*Zygomaticus major*).
  - `1 4` + `Enter`: Rapid multi-hit combo targeting both AU1 and AU4 simultaneously.
  - Formats like `1, 4` or `AU1 AU4` are also accepted.
- **Open FACS Cheat Sheet**:
  - `?` + `Enter`: View comprehensive anatomical definitions, muscle names, and movement descriptions for all Action Units.
- **Switch Language**:
  - At the title screen, enter `E` + `Enter` for English, or `J` + `Enter` for Japanese.

---

## 📂 Project Architecture

```text
├── app.db                   # SQLite3 database storing indexed stages & target AUs
├── assets/                  # Expression images & audio assets (*not tracked in git*)
│   ├── *.gif                # Practice face images (e.g., s1_4a.gif, s6_12y25.gif)
│   └── sounds/              # Sound effects (WAV) and BGM (MP3)
├── cmd/
│   ├── street-facs-fighter/ # Game entrypoint & terminal rendering engine
│   ├── sff-importer/        # Filename parsing & SQLite3 importer
│   └── gen-se/              # Mathematical WAV sound synthesizer
├── internal/
│   ├── domain/              # Entities (Question, Sample) & repository interfaces
│   ├── infrastructure/      # SQLite3 drivers, repository implementations & Sixel encoders
│   └── usecase/             # Combat loop & stage state management
├── migrations/
│   └── 001_init.sql         # Database schema for question and AU tables
├── scripts/
│   └── setup_audio.sh       # Audio placement & deployment utility
├── Makefile
└── README.md
```

---

## ⚠️ Notes & Disclaimer

- This repository **does not distribute copyrighted official FACS manual documents, test booklets, or proprietary face databases**.
- Users must supply their own authorized training images or open-source expression datasets under `assets/`.
- This software is created purely for educational, cognitive science, and anatomical training purposes.