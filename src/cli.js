import { loadConfig, saveConfig, loadLocalConfig, saveLocalConfig, findLocalConfigPath, LOCAL_FILENAME } from './config.js';
import { createWrapper, removeWrapper, syncWrappers, initShellProfile } from './shell.js';
import { generateCompletionScript } from './autocomplete.js';
import fs from 'fs';

/**
 * Shows the help message.
 */
function showHelp() {
  console.log(`
Usage: shortk <command> [args]

Commands:
  init                       Set up shell integration (PATH)
  add <short> "<long>"       Add a new short command
  remove <short>             Remove a short command
  list                       List all short commands
  status <short>             Check if command is local or global
  completion                 Generate shell completion script
  help                       Show this help message

Options:
  --local                    Apply command to local .shortk file
  `);
}

/**
 * Lists all configured aliases.
 */
function listAliases(isLocal) {
  let aliases;
  let sourceInfo = '';

  if (isLocal) {
    const localPath = findLocalConfigPath();
    if (localPath) {
      aliases = loadLocalConfig();
      sourceInfo = ` (from ${localPath})`;
    } else {
      aliases = {};
    }
  } else {
    aliases = loadConfig();
  }

  const entries = Object.entries(aliases);

  if (entries.length === 0) {
    console.log(`No ${isLocal ? 'local ' : ''}short commands configured yet${isLocal && !sourceInfo ? ' (no .shortk file found)' : ''}.`);
    return;
  }

  console.log(`Your ${isLocal ? 'local ' : ''}short commands${sourceInfo}:`);
  entries.forEach(([short, long]) => {
    console.log(`  ${short.padEnd(15)} ->  ${long}`);
  });
}

/**
 * Adds a new alias.
 */
function addAlias(short, long, isLocal) {
  if (!short || !long) {
    console.error('Error: Both <short> and <long> commands are required.');
    return;
  }

  if (isLocal) {
    const aliases = loadLocalConfig();
    aliases[short] = long;
    saveLocalConfig(aliases);
    
    // Create global wrapper if it doesn't exist, so the binary is in PATH
    const globalAliases = loadConfig();
    if (!globalAliases[short]) {
      createWrapper(short, ""); 
    }
    console.log(`Successfully added LOCAL: ${short} -> ${long} (Saved to ${LOCAL_FILENAME})`);
  } else {
    const aliases = loadConfig();
    aliases[short] = long;
    saveConfig(aliases);
    
    if (createWrapper(short, long)) {
      console.log(`Successfully added GLOBAL: ${short} -> ${long}`);
    }
  }
}

/**
 * Removes an alias.
 */
function removeAlias(short, isLocal) {
  if (!short) {
    console.error('Error: <short> command is required.');
    return;
  }

  if (isLocal) {
    const aliases = loadLocalConfig();
    if (!aliases[short]) {
      console.error(`Error: Local short command "${short}" not found in .shortk`);
      return;
    }
    delete aliases[short];
    saveLocalConfig(aliases);
    console.log(`Successfully removed LOCAL: ${short}`);
  } else {
    const aliases = loadConfig();
    if (!aliases[short]) {
      console.error(`Error: Global short command "${short}" not found.`);
      return;
    }
    delete aliases[short];
    saveConfig(aliases);
    
    // Check if it exists in local config before removing wrapper?
    // Actually syncWrappers() will handle it.
    removeWrapper(short);
    console.log(`Successfully removed GLOBAL: ${short}`);
  }
}

/**
 * Checks the status of a short command.
 */
function checkStatus(short) {
  if (!short) {
    console.error('Error: <short> command is required.');
    return;
  }

  const localPath = findLocalConfigPath();
  let localCmd = null;
  if (localPath) {
    const localConfig = {};
    const data = fs.readFileSync(localPath, 'utf-8');
    data.split('\n').forEach(line => {
      const idx = line.indexOf('=');
      if (idx > 0) localConfig[line.substring(0, idx).trim()] = line.substring(idx + 1).trim();
    });
    localCmd = localConfig[short];
  }

  const globalConfig = loadConfig();
  const globalCmd = globalConfig[short];

  if (!localCmd && !globalCmd) {
    console.log(`'${short}' is not defined.`);
    return;
  }

  console.log(`Status for '${short}':`);
  if (localCmd) {
    console.log(`  LOCAL:  ${localCmd} (from ${localPath})`);
    if (globalCmd) {
      console.log(`  GLOBAL: ${globalCmd} (OVERRIDDEN)`);
    }
  } else {
    console.log(`  GLOBAL: ${globalCmd}`);
    console.log(`  LOCAL:  (none found in directory tree)`);
  }
}

/**
 * Main entry point for the CLI.
 */
export async function run(args) {
  const isLocal = args.includes('--local');
  const cleanArgs = args.filter(a => a !== '--local');
  const [command, ...rest] = cleanArgs;

  switch (command) {
    case 'init':
      if (initShellProfile()) {
        syncWrappers();
      }
      break;
    case 'add':
      addAlias(rest[0], rest[1], isLocal);
      break;
    case 'remove':
      removeAlias(rest[0], isLocal);
      break;
    case 'list':
      listAliases(isLocal);
      break;
    case 'status':
      checkStatus(rest[0]);
      break;
    case '_list-keys':
      try {
        const globalKeys = Object.keys(loadConfig());
        let localKeys = [];
        const localPath = findLocalConfigPath();
        if (localPath) {
          const data = fs.readFileSync(localPath, 'utf-8');
          data.split('\n').forEach(line => {
            const idx = line.indexOf('=');
            if (idx > 0) localKeys.push(line.substring(0, idx).trim());
          });
        }
        const allKeys = [...new Set([...globalKeys, ...localKeys])].filter(k => k);
        process.stdout.write(allKeys.join(' '));
      } catch (e) {
        // Silent fail for autocomplete
      }
      break;
    case 'completion':
      console.log(generateCompletionScript());
      break;
    case 'help':
    case '--help':
    case '-h':
      showHelp();
      break;
    default:
      if (command) {
        console.error(`Unknown command: ${command}`);
      }
      showHelp();
      break;
  }
}
