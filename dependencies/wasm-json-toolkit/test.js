const fs = require('fs')
const Iter = require('./iterator.js')
const wasm = fs.readFileSync(`./test/customSection.wasm`)
const it = new Iter(wasm)
for (const section of it) {
  console.log(section.type)
}
