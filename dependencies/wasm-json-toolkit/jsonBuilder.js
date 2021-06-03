const leb128 = require('leb128').unsigned

exports.i32 = {
  // memory loads
  load8S: function (offset, flags) {
    return {
      "return_type": "i32",
      "name": "load8_s",
      "immediates": {
        flags,
        offset
      }
    }
  },
  load8U: function (offset, flags) {
    return {
      "return_type": "i32",
      "name": "load8_u",
      "immediates": {
        flags,
        offset
      }
    }
  },
  load16S: function () {
    return {
      "return_type": "i32",
      "name": "load16_s",
      "immediates": {
        flags,
        offset
      }
    }
  },
  load16U: function () {},
  load: function () {},

  // memory store
  store8: function () {},
  store16: function () {},
  store: function () {},

  // const
  const: function (immediate) {
    return {
      'return_type': 'i32',
      'name': 'const',
      'immediates': immediate
    }
  },

  // Integer operators
  // sign-agnostic addition
  add: function () {},
  // sign-agnostic subtraction
  sub: function () {},
  // sign-agnostic multiplication (lower 32-bits)
  mul: function () {},
  // signed division (result is truncated toward zero)
  divS: function () {},
  // unsigned division (result is floored)
  divU: function () {},
  // signed remainder (result has the sign of the dividend)
  remS: function () {},
  // unsigned remainder
  remU: function () {},
  // sign-agnostic bitwise and
  and: function () {},
  // sign-agnostic bitwise inclusive or
  or: function () {},
  // sign-agnostic bitwise exclusive or
  xor: function () {},
  // sign-agnostic shift left
  shl: function () {},
  // zero-replicating (logical) shift right
  shrU: function () {},
  // sign-replicating (arithmetic) shift right
  shrS: function () {},
  // sign-agnostic rotate left
  rotl: function () {},
  // sign-agnostic rotate right
  rotr: function () {},
  // sign-agnostic compare equal
  eq: function () {},
  // sign-agnostic compare unequal
  ne: function () {},
  // signed less than
  ltS: function () {},
  // signed less than or equal
  leS: function () {},
  // unsigned less than
  ltU: function () {},
  // unsigned less than or equal
  leU: function () {},
  // signed greater than
  gtS: function () {},
  // signed greater than or equal
  geS: function () {},
  // unsigned greater than
  gtU: function () {},
  // unsigned greater than or equal
  geU: function () {},
  // sign-agnostic count leading zero bits (All zero bits are considered leading if the value is zero)
  clz: function () {},
  // sign-agnostic count trailing zero bits (All zero bits are considered trailing if the value is zero)
  ctz: function () {},
  // sign-agnostic count number of one bits
  popcnt: function () {},
  // compare equal to zero (return 1 if operand is zero, 0 otherwise)
  eqz: function () {},
  // wrap a 64-bit integer to a 32-bit integer
  wrapI64: function () {},
  // truncate a 32-bit float to a signed 32-bit integer
  truncSF32: function () {},
  // truncate a 64-bit float to a signed 32-bit integer
  truncSF64: function () {},
  // truncate a 32-bit float to an unsigned 32-bit integer
  truncUF32: function () {},
  // truncate a 64-bit float to an unsigned 32-bit integer
  truncUF64: function () {},
  // reinterpret the bits of a 32-bit float as a 32-bit integer
  reinterpretF32: function () {}
}

exports.i64 = {
  // memory loads
  load8S: function () {},
  load8U: function () {},
  load16S: function () {},
  load16U: function () {},
  load32S: function () {},
  load32U: function () {},
  load: function () {},

  // memory store
  store8: function () {},
  store16: function () {},
  store32: function () {},
  store: function () {},

  const: function (immediate) {
    return {
      'return_type': 'i64',
      'name': 'const',
      'immediates': immediate
    }
  },

  // Integer operators
  // sign-agnostic addition
  add: function () {},
  // sign-agnostic subtraction
  sub: function () {},
  // sign-agnostic multiplication (lower 32-bits)
  mul: function () {},
  // signed division (result is truncated toward zero)
  divS: function () {},
  // unsigned division (result is floored)
  divU: function () {},
  // signed remainder (result has the sign of the dividend)
  remS: function () {},
  // unsigned remainder
  remU: function () {},
  // sign-agnostic bitwise and
  and: function () {},
  // sign-agnostic bitwise inclusive or
  or: function () {},
  // sign-agnostic bitwise exclusive or
  xor: function () {},
  // sign-agnostic shift left
  shl: function () {},
  // zero-replicating (logical) shift right
  shrU: function () {},
  // sign-replicating (arithmetic) shift right
  shrS: function () {},
  // sign-agnostic rotate left
  rotl: function () {},
  // sign-agnostic rotate right
  rotr: function () {},
  // sign-agnostic compare equal
  eq: function () {},
  // sign-agnostic compare unequal
  ne: function () {},
  // signed less than
  ltS: function () {},
  // signed less than or equal
  leS: function () {},
  // unsigned less than
  ltU: function () {},
  // unsigned less than or equal
  leU: function () {},
  // signed greater than
  gtS: function () {},
  // signed greater than or equal
  geS: function () {},
  // unsigned greater than
  gtU: function () {},
  // unsigned greater than or equal
  geU: function () {},
  // sign-agnostic count leading zero bits (All zero bits are considered leading if the value is zero)
  clz: function () {},
  // sign-agnostic count trailing zero bits (All zero bits are considered trailing if the value is zero)
  ctz: function () {},
  // sign-agnostic count number of one bits
  popcnt: function () {},
  // compare equal to zero (return 1 if operand is zero, 0 otherwise)
  eqz: function () {},
  /// extend a signed 32-bit integer to a 64-bit integer
  extendSI32: function () {},
  // extend an unsigned 32-bit integer to a 64-bit integer
  extendUI32: function () {},
  // truncate a 32-bit float to a signed 64-bit integer
  truncSF32: function () {},
  // truncate a 64-bit float to a signed 64-bit integer
  truncSF64: function () {},
  // truncate a 32-bit float to an unsigned 64-bit integer
  truncUF32: function () {},
  // truncate a 64-bit float to an unsigned 64-bit integer
  truncUF64: function () {},
  // reinterpret the bits of a 64-bit float as a 64-bit integer
  reinterpretF64: function () {}
  }

exports.f32 = {
  // memory loads
  load: function () {},
  // memory store
  store: function () {},
  const: function (immediate) {
    return {
      'return_type': 'f32',
      'name': 'const',
      'immediates': immediate
    }
  },

  // Integer operators
  // sign-agnostic addition
  add: function () {},
  // sign-agnostic subtraction
  sub: function () {},
  // sign-agnostic multiplication (lower 32-bits)
  mul: function () {},
  // unsigned division 
  div: function () {},
  // absolute value
  abs: function () {},
  // negation
  neg: function () {},
  // copysign
  copysign: function () {},
  // ceiling operator
  ceil: function () {},
  // floor operator
  floor: function () {},
  // round to nearest integer towards zero
  trunc: function () {},
  // round to nearest integer, ties to even
  nearest: function () {},
  // compare ordered and equal
  eq: function () {},
  // compare unordered or unequal
  ne: function () {},
  // compare ordered and less than
  lt: function () {},
  // compare ordered and less than or equal
  le: function () {},
  // compare ordered and greater than
  gt: function () {},
  // compare ordered and greater than or equal
  ge: function () {},
  // square root
  sqrt: function () {},
  // minimum (binary operator); if either operand is NaN, returns NaN
  min: function () {},
  // maximum (binary operator); if either operand is NaN, returns NaN
  max: function () {},

  // Datatype conversions, truncations, reinterpretations, promotions, and demotions
  // demote a 64-bit float to a 32-bit float
  demoteF64: function () {},
  // convert a signed 32-bit integer to a 32-bit float
  convertSI32: function () {},
  // convert a signed 64-bit integer to a 32-bit float
  convertSI64: function () {},
  // convert an unsigned 32-bit integer to a 32-bit float
  convertUI32: function () {},
  // convert an unsigned 64-bit integer to a 32-bit float
  convertUI64: function () {},
  // reinterpret the bits of a 32-bit integer as a 32-bit float
  reinterpretI32: function () {}
}

exports.f64 = {
  // memory loads
  load: function () {},
  // memory store
  store: function () {},
  const: function (immediate) {
    return {
      'return_type': 'f64',
      'name': 'const',
      'immediates': immediate
    }
  },

  // Integer operators
  // sign-agnostic addition
  add: function () {},
  // sign-agnostic subtraction
  sub: function () {},
  // sign-agnostic multiplication (lower 32-bits)
  mul: function () {},
  // unsigned division
  div: function () {},
  // absolute value
  abs: function () {},
  // negation
  neg: function () {},
  // copysign
  copysign: function () {},
  // ceiling operator
  ceil: function () {},
  // floor operator
  floor: function () {},
  // round to nearest integer towards zero
  trunc: function () {},
  // round to nearest integer, ties to even
  nearest: function () {},
  // compare ordered and equal
  eq: function () {},
  // compare unordered or unequal
  ne: function () {},
  // compare ordered and less than
  lt: function () {},
  // compare ordered and less than or equal
  le: function () {},
  // compare ordered and greater than
  gt: function () {},
  // compare ordered and greater than or equal
  ge: function () {},
  // square root
  sqrt: function () {},
  // minimum (binary operator); if either operand is NaN, returns NaN
  min: function () {},
  // maximum (binary operator); if either operand is NaN, returns NaN
  max: function () {},

  // Datatype conversions, truncations, reinterpretations, promotions, and demotions
  // promote a 32-bit float to a 64-bit float
  promoteF32: function () {},
  // convert a signed 32-bit integer to a 64-bit float
  convertSI32: function () {},
  // convert a signed 64-bit integer to a 64-bit float
  convertSI64: function () {},
  // convert an unsigned 32-bit integer to a 64-bit float
  convertUI32: function () {},
  // convert an unsigned 64-bit integer to a 64-bit float
  convertUI64: function () {},
  // reinterpret the bits of a 64-bit integer as a 64-bit float
  reinterpretI64: function () {}
}

exports.growMemory = function () {}
exports.currentMemory = function () {}

exports.getLocal = function (index) {
  return {
    name: 'get_local',
    immediates: index
  }
}

exports.setLocal = function () {}
exports.teeLocal = function () {}

exports.setGlobals = function () {}
exports.getGlobals = function () {}

// control flow
exports.nop = function () {}
exports.block = function () {}
exports.loop = function () {}
exports.if = function () {}
exports.else = function () {}
exports.br = function () {}
exports.brIf = function () {}
exports.brTable = function () {}
exports.return = function () {}
exports.end = function () {
  return {
    name: 'end'
  }
}

exports.call = function (index) {
  return {
    'name': 'call',
    'immediates': index
  }
}

exports.callIndirect = function (index) {
  return {
    'name': 'call_indirect',
    'immediates': {
      'index': index,
      'reserved': 0
    }
  }
}

exports.drop = function () {}
exports.select = function () {}
exports.unreachable = function () {}
