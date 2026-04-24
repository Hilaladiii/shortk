import fs from 'fs';
import path from 'path';
import os from 'os';
import { CONFIG_DIR, loadConfig } from './config.js';

const BIN_DIR = path.join(CONFIG_DIR, 'bin');

/**
 * Ensures the bin directory exists.
 */
function ensureBinDir() {
  if (!fs.existsSync(BIN_DIR)) {
    fs.mkdirSync(BIN_DIR, { recursive: true });
  }
}

/**
 * Creates an executable wrapper script for a command.
 * Uses a Smart Wrapper that checks for local .shortk overrides.
 */
export function createWrapper(short, long) {
  ensureBinDir();
  const scriptPath = path.join(BIN_DIR, short);
  
  // Escape single quotes for the bash script
  const escapedLong = (long || '').replace(/'/g, "'\\''");
  
  // Bash script that checks for local .shortk file in current or parent dirs
  const content = `#!/usr/bin/env bash

SHORT="${short}"
GLOBAL_CMD='${escapedLong}'

load_env() {
  if [ -f "$1" ]; then
    set -a
    source "$1"
    set +a
  fi
}

# Traverse up to find .shortk
dir="$PWD"
while [ "$dir" != "/" ] && [ "$dir" != "." ]; do
  if [ -f "$dir/.shortk" ]; then
    # Use awk to find the exact match key=value
    LOCAL_CMD=$(awk -F'=' -v k="$SHORT" '$1==k {sub(/^[^=]+=/,""); print; exit}' "$dir/.shortk")
    if [ -n "$LOCAL_CMD" ]; then
      load_env "$dir/.env"
      eval "$LOCAL_CMD \\"\$@\\""
      exit $?
    fi
  fi
  dir=$(dirname "$dir")
done

if [ -n "$GLOBAL_CMD" ] && [ "$GLOBAL_CMD" != "undefined" ]; then
  load_env "$PWD/.env"
  eval "$GLOBAL_CMD \\"\$@\\""
else
  echo "shortk: '$SHORT' is not defined globally and no local override found."
  exit 1
fi
`;
  
  try {
    fs.writeFileSync(scriptPath, content, 'utf-8');
    fs.chmodSync(scriptPath, 0o755); // Make it executable
    return true;
  } catch (error) {
    console.error(`Error creating wrapper for ${short}:`, error.message);
    return false;
  }
}

/**
 * Removes an executable wrapper script.
 */
export function removeWrapper(short) {
  const scriptPath = path.join(BIN_DIR, short);
  if (fs.existsSync(scriptPath)) {
    try {
      fs.unlinkSync(scriptPath);
      return true;
    } catch (error) {
      console.error(`Error removing wrapper for ${short}:`, error.message);
      return false;
    }
  }
  return true;
}

/**
 * Re-generates all wrappers based on config.
 */
export function syncWrappers() {
  ensureBinDir();
  
  // Clear existing wrappers
  if (fs.existsSync(BIN_DIR)) {
    const files = fs.readdirSync(BIN_DIR);
    for (const file of files) {
      fs.unlinkSync(path.join(BIN_DIR, file));
    }
  }
  
  // Create wrappers from config
  const aliases = loadConfig();
  for (const [short, long] of Object.entries(aliases)) {
    createWrapper(short, long);
  }
}

/**
 * Adds the PATH export to the detected shell profile.
 */
export function initShellProfile() {
  const profiles = [
    path.join(os.homedir(), '.zshrc'),
    path.join(os.homedir(), '.bashrc')
  ].filter(p => fs.existsSync(p));

  if (profiles.length === 0) {
    console.error('No shell profile found (e.g., .zshrc, .bashrc).');
    return false;
  }

  const startMarker = '# <<< shortk initialize <<<';
  const endMarker = '# >>> shortk initialize >>>';
  
  const binPath = path.join(os.homedir(), '.config', 'shortk', 'bin');
  const integrationCode = [
    startMarker,
    `export PATH="${binPath}:$PATH"`,
    `source <(shortk completion)`,
    endMarker
  ].join('\n');
  
  for (const profile of profiles) {
    try {
      let content = fs.readFileSync(profile, 'utf-8');
      
      // Remove old alias-based integration if exists
      content = content.replace(/# shortk configuration[\s\S]*?shortk\.sh" # added by shortk\n*/g, '');
      
      if (content.includes(startMarker)) {
        const regex = new RegExp(`${startMarker}[\\s\\S]*?${endMarker}`, 'g');
        content = content.replace(regex, integrationCode);
        fs.writeFileSync(profile, content, 'utf-8');
        console.log(`Updated integration in ${profile}`);
      } else {
        fs.appendFileSync(profile, `\n\n${integrationCode}\n`, 'utf-8');
        console.log(`Successfully added integration to ${profile}`);
      }
    } catch (error) {
      console.error(`Error updating ${profile}:`, error.message);
    }
  }
  
  console.log('\nIMPORTANT: To apply changes, please restart your terminal or run:');
  profiles.forEach(p => console.log(`source ${p}`));
  
  return true;
}
