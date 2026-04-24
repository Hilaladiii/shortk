import fs from 'fs';
import path from 'path';
import os from 'os';

const CONFIG_DIR = path.join(os.homedir(), '.config', 'shortk');
const ALIASES_FILE = path.join(CONFIG_DIR, 'aliases.json');
const LOCAL_FILENAME = '.shortk';

/**
 * Ensures the configuration directory exists.
 */
function ensureConfigDir() {
  if (!fs.existsSync(CONFIG_DIR)) {
    fs.mkdirSync(CONFIG_DIR, { recursive: true });
  }
}

/**
 * Loads the aliases from the global configuration file.
 * @returns {Object} An object mapping short commands to long commands.
 */
export function loadConfig() {
  ensureConfigDir();
  if (!fs.existsSync(ALIASES_FILE)) {
    return {};
  }
  try {
    const data = fs.readFileSync(ALIASES_FILE, 'utf-8');
    return JSON.parse(data);
  } catch (error) {
    console.error('Error loading configuration:', error.message);
    return {};
  }
}

/**
 * Saves the aliases to the global configuration file.
 * @param {Object} config - An object mapping short commands to long commands.
 */
export function saveConfig(config) {
  ensureConfigDir();
  try {
    fs.writeFileSync(ALIASES_FILE, JSON.stringify(config, null, 2), 'utf-8');
  } catch (error) {
    console.error('Error saving configuration:', error.message);
  }
}

/**
 * Loads the local aliases from the nearest .shortk file.
 * @returns {Object} An object mapping short commands to long commands.
 */
export function loadLocalConfig() {
  const localPath = findLocalConfigPath();
  if (!localPath) {
    return {};
  }
  try {
    const data = fs.readFileSync(localPath, 'utf-8');
    const lines = data.split('\n');
    const config = {};
    for (const line of lines) {
      const index = line.indexOf('=');
      if (index > 0) {
        const key = line.substring(0, index).trim();
        const value = line.substring(index + 1).trim();
        if (key) config[key] = value;
      }
    }
    return config;
  } catch (error) {
    console.error('Error loading local configuration:', error.message);
    return {};
  }
}

/**
 * Saves the local aliases to the nearest .shortk file or creates a new one in cwd.
 * @param {Object} config - An object mapping short commands to long commands.
 */
export function saveLocalConfig(config) {
  const localPath = findLocalConfigPath() || path.join(process.cwd(), LOCAL_FILENAME);
  try {
    const data = Object.entries(config)
      .map(([key, value]) => `${key}=${value}`)
      .join('\n');
    fs.writeFileSync(localPath, data, 'utf-8');
    addToGitIgnore(localPath);
  } catch (error) {
    console.error('Error saving local configuration:', error.message);
  }
}

/**
 * Adds .shortk to .gitignore if it's not already there.
 * @param {string} shortkPath - Path to the .shortk file.
 */
function addToGitIgnore(shortkPath) {
  const dir = path.dirname(shortkPath);
  const gitignorePath = path.join(dir, '.gitignore');
  try {
    if (fs.existsSync(gitignorePath)) {
      const content = fs.readFileSync(gitignorePath, 'utf-8');
      if (!content.includes(LOCAL_FILENAME)) {
        fs.appendFileSync(gitignorePath, `\n${LOCAL_FILENAME}\n`, 'utf-8');
      }
    } else {
      // Create .gitignore if in a git repo
      if (fs.existsSync(path.join(dir, '.git'))) {
        fs.writeFileSync(gitignorePath, `${LOCAL_FILENAME}\n`, 'utf-8');
      }
    }
  } catch (error) {
    // Ignore errors for gitignore
  }
}

/**
 * Finds the nearest .shortk file by traversing up the directory tree.
 * @returns {string|null} Absolute path to the nearest .shortk file or null.
 */
export function findLocalConfigPath() {
  let currentDir = process.cwd();
  while (currentDir !== path.parse(currentDir).root) {
    const localPath = path.join(currentDir, LOCAL_FILENAME);
    if (fs.existsSync(localPath)) {
      return localPath;
    }
    currentDir = path.dirname(currentDir);
  }
  // Check root
  const rootPath = path.join(path.parse(currentDir).root, LOCAL_FILENAME);
  if (fs.existsSync(rootPath)) return rootPath;
  
  return null;
}

export { CONFIG_DIR, ALIASES_FILE, LOCAL_FILENAME };
