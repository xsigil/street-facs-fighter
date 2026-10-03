# 🥊 Street FACS Fighter

> **The Ultimate Action Unit Battle Arena in Your Terminal**  
> Master facial muscle decoding (FACS) through real-time ATB combat and high-resolution Sixel graphics in a retro arcade-style terminal fighting game!

---

## 🌟 Overview

**Street FACS Fighter** is a terminal-based arcade combat game themed around the **Facial Action Coding System (FACS)**—the anatomical and psychological standard for categorizing human facial expressions.

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

### 3. Import Expression Dataset

Place your licensed FACS training images (e.g., named with AU identifiers such as `s1_4.gif` or `s12.jpg`) into `assets/`, then build the SQLite index:

```bash
make import
```

### 4. Configure BGM Tracks (Optional)

Place `assets/sounds/stage1.mp3` (combat background music) and `assets/sounds/lose.mp3` (results screen theme) to enable arcade soundtrack playback.

### 5. Launch the Game

```bash
# Launch in English mode directly
go run ./cmd/street-facs-fighter -lang en

# Or launch with default settings
make run
```

---

## 🎮 How to Play

- **Strike Weak-Point AUs**:
  - `12` + `Enter`: Single strike targeting AU12 (*Zygomaticus major*).
  - `1 4` + `Enter`: Rapid multi-hit combo targeting both AU1 and AU4.
- **Open FACS Cheat Sheet**:
  - `?` + `Enter`: View comprehensive anatomical definitions and movement descriptions for all Action Units.
- **Switch Language**:
  - At the title screen, enter `E` + `Enter` for English, or `J` + `Enter` for Japanese.

---

## 📂 Project Architecture

```text
├── assets/                  # Expression images & audio assets (*not tracked in git*)
├── cmd/
│   ├── street-facs-fighter/ # Game entrypoint & terminal rendering engine
│   ├── gen-se/              # Mathematical WAV sound synthesizer
│   └── import/              # Image filename parsing & SQLite3 importer
├── internal/
│   ├── domain/              # Entities & repository interfaces
│   ├── infrastructure/      # SQLite3 drivers & Sixel encoders
│   └── usecase/             # Combat loop & stage state management
├── scripts/
│   └── setup_audio.sh       # Audio placement & deployment utility
├── Makefile
└── README.md
```

---

## ⚠️ Notes & Disclaimer

- This repository **does not contain copyrighted official FACS manual documents, test booklets, or proprietary face databases**.
- Users must supply their own authorized training images or open-source expression datasets under `assets/`.
- This software is created purely for educational, cognitive science, and anatomical training purposes.