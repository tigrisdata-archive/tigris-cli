#!/usr/bin/env node

/* eslint-disable no-console */

const fetch = require('node-fetch'); const path = require('path');
const tar = require('tar'); const mkdirp = require('mkdirp');
const fs = require('fs'); const { execSync } = require('child_process');
const crypto = require('crypto');

// Mapping from Node's `process.arch` to Golang's `$GOARCH`
const ARCH_MAPPING = {
  ia32: '386',
  x64: 'amd64',
  arm: 'arm64',
};

// Mapping between Node's `process.platform` to Golang's
const PLATFORM_MAPPING = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
  freebsd: 'freebsd',
};

function getInstallationPath(callback) {
  const out = execSync('npm root');

  let dir = null;
  if (out.length === 0) {
    console.error("couldn't determine executable path");
  } else {
    dir = `${out.toString().trim()}/.bin`;

    console.log(`Installing tigris binary to:  ${dir}`);

    mkdirp.sync(dir);

    callback(null, dir);
  }
}

function verifyAndPlaceBinary(binName, binPath, hash, checksums, callback) {
  if (!fs.existsSync(path.join(binPath, binName))) {
    return callback(
      `Downloaded binary does not contain the binary specified in configuration - ${binPath}/${binName}`,
    );
  }

  const sum = checksums[`${PLATFORM_MAPPING[process.platform]}_${ARCH_MAPPING[process.arch]}`];
  if (sum === undefined || sum !== hash) {
    return callback(`cannot validate checksum of the downloaded package. got ${hash}, expected ${sum}`);
  }

  getInstallationPath((err, installationPath) => {
    if (err) return callback('Error getting binary installation path from `npm bin`');

    // Move the binary file
    fs.renameSync(
      path.join(binPath, binName),
      path.join(installationPath, binName),
    );

    return null;
  });

  return callback(null);
}

function validateConfiguration(packageJson) {
  if (!packageJson.version) {
    return 'version property must be specified';
  }

  if (!packageJson.goBinary || typeof (packageJson.goBinary) !== 'object') {
    return 'goBinary property must be defined and be an object';
  }

  if (!packageJson.goBinary.name) {
    return 'name property is necessary';
  }

  if (!packageJson.goBinary.url) {
    return 'url property is required';
  }

  if (!packageJson.bin || typeof (packageJson.bin) !== 'object') {
    return 'bin property of package.json must be defined and be an object';
  }

  return null;
}

function parsePackageJson() {
  if (!(process.arch in ARCH_MAPPING)) {
    console.error(`Installation is not supported for this architecture: ${process.arch}`);
    return null;
  }

  if (!(process.platform in PLATFORM_MAPPING)) {
    console.error(`Installation is not supported for this platform: ${process.platform}`);
    return null;
  }

  const packageJsonPath = path.join('.', 'package.json');
  if (!fs.existsSync(packageJsonPath)) {
    console.error(
      'Unable to find package.json. '
            + 'Please run this script at root of the package you want to be installed',
    );
    return null;
  }

  const packageJson = JSON.parse(fs.readFileSync(packageJsonPath));
  const error = validateConfiguration(packageJson);
  if (error && error.length > 0) {
    console.error(`Invalid package.json: ${error}`);
    return null;
  }

  // We have validated the config. It exists in all its glory
  let { url } = packageJson.goBinary;
  let { version } = packageJson;
  if (version[0] === 'v') version = version.substr(1); // strip the 'v' if necessary v0.0.1 => 0.0.1

  const opts = {
    binName: packageJson.goBinary.name,
    binPath: '/tmp',
    version,
    checksums: packageJson.goBinary.checksums,
  };

  // Binary name on Windows has .exe suffix
  if (process.platform === 'win32') {
    opts.binName += '.exe';
  }

  // Interpolate variables in URL, if necessary
  url = url.replace(/{{arch}}/g, ARCH_MAPPING[process.arch]);
  url = url.replace(/{{platform}}/g, PLATFORM_MAPPING[process.platform]);
  url = url.replace(/{{version}}/g, version);
  url = url.replace(/{{bin_name}}/g, opts.binName);

  opts.url = url;

  return opts;
}

/**
 * Reads the configuration from application's package.json,
 * validates properties, downloads the binary, untars, and stores at
 * ./bin in the package's root. NPM already has support to install binary files
 * specific locations when invoked with 'npm install -g'
 *
 *  See: https://docs.npmjs.com/files/package.json#bin
 */
const INVALID_INPUT = 'Invalid inputs';
function install(callback) {
  const opts = parsePackageJson();
  if (!opts) return callback(INVALID_INPUT);

  mkdirp.sync(opts.binPath);

  console.log(`Downloading from URL: ${opts.url}`);
  fetch(opts.url).then((res) => {
    const hash = crypto.createHash('sha256').setEncoding('hex');
    res.body.pipe(hash).on('end', () => { hash.end(); });
    res.body
      .pipe(
        tar.x(
          {
            C: opts.binPath,
          },
          ['tigris'],
        ),
      )
      .on('end', () => {
        verifyAndPlaceBinary(
          opts.binName,
          opts.binPath,
          hash.read(),
          opts.checksums,
          callback,
        );
      });
  });

  return null;
}

function uninstall(callback) {
  const opts = parsePackageJson();
  getInstallationPath((err, installationPath) => {
    if (err) callback('Error finding binary installation directory');

    try {
      fs.unlinkSync(path.join(installationPath, opts.binName));
    } catch (ex) {
      // Ignore errors when deleting the file.
    }

    return callback(null);
  });
}

// Parse command line arguments and call the right method
const actions = {
  install,
  uninstall,
};

const { argv } = process;

if (argv && argv.length > 2) {
  let cmd = process.argv[2];
  if (cmd === undefined) {
    cmd = 'install';
  } else if (!actions[cmd]) {
    console.log(
      'Invalid command to go-npm. `install` and `uninstall` are the only supported commands',
    );
    process.exit(1);
  }

  actions[cmd]((err) => {
    if (err) {
      console.error(err);
      process.exit(1);
    } else {
      process.exit(0);
    }
  });
}
