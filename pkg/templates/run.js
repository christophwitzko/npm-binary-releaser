#!/usr/bin/env node

const os = require('os')
const pkg = require('./package.json')

const binFile = require.resolve((pkg.binPkgPrefix || '') + pkg.name + '-' + os.platform() + '-' + os.arch())

require('child_process').spawn(binFile, process.argv.slice(2), {
  cwd: process.cwd(),
  stdio: 'inherit'
})
