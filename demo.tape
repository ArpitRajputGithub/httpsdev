# VHS tape → demo.gif
# Install VHS (one time):  brew install vhs
# Render:                  vhs demo.tape

Output demo.gif

# ── canvas ────────────────────────────────────
Set FontSize 20
Set Width 1400
Set Height 800
Set Padding 30
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set Framerate 30
Set PlaybackSpeed 1.0

# ── scene ─────────────────────────────────────
# Assumes a dummy dev server is already running on :5173 in a
# separate process. Easiest: `python3 -m http.server 5173`.

Type "httpsdev --tui 5173"
Sleep 500ms
Enter
Sleep 2s

# Simulate a few incoming requests to make the feed populate.
# Requires a helper that hits localhost:3443 while VHS records —
# either open a browser tab manually or run in another shell:
#   for i in {1..6}; do curl -sk https://localhost:3443/ > /dev/null; sleep 0.5; done

Sleep 8s

# Wave goodbye
Type "q"
Sleep 500ms
