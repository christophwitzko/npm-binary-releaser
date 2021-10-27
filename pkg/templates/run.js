#!/usr/bin/env node

const process = require('process')
const cp = require('child_process')
const pkg = require('./package.json')

const binFile = require.resolve((pkg.binPkgPrefix || '') + pkg.name + '-' + process.platform + '-' + process.arch)

const subprocess = cp.spawn(binFile, process.argv.slice(2), {
  cwd: process.cwd(),
  stdio: 'inherit'
})

;[
  'SIGTERM',
  'SIGINT',
  'SIGQUIT',
  'SIGHUP',
  'SIGUSR1',
  'SIGUSR2',
  'SIGPIPE',
  'SIGBREAK',
  'SIGWINCH'
].forEach(sig => {
  process.on(sig, () => subprocess.kill(sig))
})

subprocess.on('close', (code) => process.exit(code === 0 ? 0 : code || 1))
