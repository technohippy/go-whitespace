// Whitespace Interpreter
// http://compsoc.dur.ac.uk/whitespace/
package main

import (
  "bufio"
  "flag"
  "fmt"
  "os"
  "strconv"
)

const (
  NULL_CHAR = -1
  STACK_SIZE = 30000
)

const (
  STAGE_IMP = iota
  STAGE_instruction
  STAGE_PARAM

  IMP_STACK
  IMP_ARITHMETIC
  IMP_HEAP
  IMP_FLOW
  IMP_IO

  STACK_PUSH
  STACK_DUPLICATE
  STACK_COPY
  STACK_SWAP
  STACK_DISCARD
  STACK_SLIDE

  ARITHMETIC_ADD
  ARITHMETIC_SUB
  ARITHMETIC_MUL
  ARITHMETIC_DIV
  ARITHMETIC_MOD

  HEAP_STORE
  HEAP_RETRIEVE

  FLOW_MARK
  FLOW_CALL_SUB
  FLOW_JUMP
  FLOW_JUMP_ZERO
  FLOW_JUMP_NEGATIVE
  FLOW_END_SUB
  FLOW_END

  IO_OUT_CHAR
  IO_OUT_NUM
  IO_IN_CHAR
  IO_IN_NUM
)

var is_debug bool
var last_char = NULL_CHAR

var current_stage = STAGE_IMP
var current_imp int
var current_instruction int

var program [STACK_SIZE][3]int
var pc = 0

var marks = make(map[int] int)

var heap = make(map[int] int)
var stack [STACK_SIZE]int
var sc = 0

var call_stack [STACK_SIZE]int
var csc = 0

/*
 * Parse
 */
func parse(file *os.File) {
  reader := bufio.NewReader(file)
  for {
    line, err := reader.ReadString('\n')
    if err != nil {
      if err == os.EOF {
        break;
      } else {
        error(err.String())
      }
    }

    for _, c := range line {
      switch c {
      case ' ' , '\t', '\r', '\n':
        switch current_stage {
        case STAGE_IMP: parse_imp(c)
        case STAGE_instruction: parse_instruction(c)
        case STAGE_PARAM: parse_parameter(c)
        }
      }
    }
  }
}

// Instruction Modification Parameter
func parse_imp(c int) {
  switch c {
  case ' ':
    switch last_char {
    case NULL_CHAR:
      // [Space]
      debug("[IMP:STACK] ")
      current_imp = IMP_STACK
      set_stage(STAGE_instruction)
    case '\t':
      // [Tab][Space]
      debug("[IMP:ARITHMETIC] ")
      current_imp = IMP_ARITHMETIC
      set_stage(STAGE_instruction)
    }
  case '\t':
    switch last_char {
    case '\t':
      // [Tab][Tab]
      debug("[IMP:HEAP] ")
      current_imp = IMP_HEAP
      set_stage(STAGE_instruction)
    }
  case '\r', '\n':
    switch last_char {
    case NULL_CHAR:
      // [LF]
      debug("[IMP:FLOW] ")
      current_imp = IMP_FLOW
      set_stage(STAGE_instruction)
    case '\t':
      // [Tab][LF]
      debug("[IMP:I/O] ")
      current_imp = IMP_IO
      set_stage(STAGE_instruction)
    }
  }
  if current_stage == STAGE_IMP {
    last_char = c
  }
}

// Instructions
func parse_instruction(c int) {
  switch current_imp {
  case IMP_STACK:
    parse_instruction_stack(c)
  case IMP_ARITHMETIC:
    parse_instruction_arithmetic(c)
  case IMP_HEAP:
    parse_instruction_heap(c)
  case IMP_FLOW:
    parse_instruction_flow(c)
  case IMP_IO:
    parse_instruction_io(c)
  }
  if current_stage == STAGE_instruction {
    last_char = c
  }
}

// Stack Manipulation (IMP: [Space])
func parse_instruction_stack(c int) {
  switch c {
  case ' ':
    switch last_char {
    case NULL_CHAR:
      // [Space] Number
      debug("[Number] ")
      current_instruction = STACK_PUSH
      set_stage(STAGE_PARAM)
    case '\r', '\n':
      // [LF][Space]
      debug("[Duplicate]\n")
      push_instruction([3]int{STACK_DUPLICATE, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Space]
      debug("[Copy] ")
      current_instruction = STACK_COPY
      set_stage(STAGE_PARAM)
    }
  case '\t':
    switch last_char {
    case '\r', '\n':
      // [LF][Tab]
      debug("[Copy]\n")
      push_instruction([3]int{STACK_COPY, 0, 0})
      set_stage(STAGE_IMP)
    default:
      error(fmt.Sprintf("Parse Error: stack: '%c'", c))
    }
  case '\r', '\n':
    switch last_char {
    case NULL_CHAR:
      // do nothing
    case '\r', '\n':
      // [LF][LF]
      debug("[Discard]\n")
      push_instruction([3]int{STACK_DISCARD, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][LF] Number
      debug("[Slide] ")
      current_instruction = STACK_SLIDE
      set_stage(STAGE_PARAM)
    default:
      error(fmt.Sprintf("Parse Error: stack: '%c'", c))
    }
  }
}

// Arithmetic (IMP: [Tab][Space])
func parse_instruction_arithmetic(c int) {
  switch c {
  case ' ':
    switch last_char {
    case ' ':
      // [Space][Space]
      debug("[Add]\n")
      push_instruction([3]int{ARITHMETIC_ADD, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Space]
      debug("[Devision]\n")
      push_instruction([3]int{ARITHMETIC_DIV, 0, 0})
      set_stage(STAGE_IMP)
    }
  case '\t':
    switch last_char {
    case ' ':
      // [Space][Tab]
      debug("[Subtraction]\n")
      push_instruction([3]int{ARITHMETIC_SUB, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Tab]
      debug("[Modulo]\n")
      push_instruction([3]int{ARITHMETIC_MOD, 0, 0})
      set_stage(STAGE_IMP)
    }
  case '\r', '\n':
    switch last_char {
    case ' ':
      // [Space][LF]
      debug("[Multiplication]\n")
      push_instruction([3]int{ARITHMETIC_MUL, 0, 0})
      set_stage(STAGE_IMP)
    default:
      error(fmt.Sprintf("Parse Error: arithmetic: '%c'", c))
    }
  }
}

// Heap Access (IMP: [Tab][Tab])
func parse_instruction_heap(c int) {
  switch c {
  case ' ':
    // [Space]
    debug("[Store]\n")
    push_instruction([3]int{HEAP_STORE, 0, 0})
    set_stage(STAGE_IMP)
  case '\t':
    // [Tab]
    debug("[Retrieve]\n")
    push_instruction([3]int{HEAP_RETRIEVE, 0, 0})
    set_stage(STAGE_IMP)
  default:
    error(fmt.Sprintf("Parse Error: heap: '%c'", c))
  }
}

// Flow Control (IMP: [LF])
func parse_instruction_flow(c int) {
  switch c {
  case ' ':
    switch last_char {
    case ' ':
      // [Space][Space] Label
      debug("[Mark] ")
      current_instruction = FLOW_MARK
      set_stage(STAGE_PARAM)
    case '\t':
      // [Tab][Space] Label
      debug("[Jump Zero] ")
      current_instruction = FLOW_JUMP_ZERO
      set_stage(STAGE_PARAM)
    }
  case '\t':
    switch last_char {
    case ' ':
      // [Space][Tab] Label
      debug("[Call] ")
      current_instruction = FLOW_CALL_SUB
      set_stage(STAGE_PARAM)
    case '\t':
      // [Tab][Tab] Label
      debug("[Jump Neg] ")
      current_instruction = FLOW_JUMP_NEGATIVE
      set_stage(STAGE_PARAM)
    }
  case '\r', '\n':
    switch last_char {
    case ' ':
      // [Space][LF] Label
      debug("[Jump] ")
      current_instruction = FLOW_JUMP
      set_stage(STAGE_PARAM)
    case '\t':
      // [Tab][LF]
      debug("[End Sub]\n")
      push_instruction([3]int{FLOW_END_SUB, 0, 0})
      set_stage(STAGE_IMP)
    case '\r', '\n':
      // [LF][LF]
      debug("[End]\n")
      push_instruction([3]int{FLOW_END, 0, 0})
      //os.Exit(0)
    }
  }
}

// I/O (IMP: [Tab][LF])
func parse_instruction_io(c int) {
  switch c {
  case ' ':
    switch last_char {
    case ' ':
      // [Space][Space]
      debug("[Out Char]\n")
      push_instruction([3]int{IO_OUT_CHAR, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Space]
      debug("[In Char]\n")
      push_instruction([3]int{IO_IN_CHAR, 0, 0})
      set_stage(STAGE_IMP)
    }
  case '\t':
    switch last_char {
    case ' ':
      // [Space][Space]
      debug("[Out Num]\n")
      push_instruction([3]int{IO_OUT_NUM, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Space]
      debug("[In Num]\n")
      push_instruction([3]int{IO_IN_NUM, 0, 0})
      set_stage(STAGE_IMP)
    default:
      //error(fmt.Sprintf("io error 2: '%c'", last_char))
    }
  }
}

var sign = 0
var abs = 0

// Number, Label
func parse_parameter(c int) {
  if sign == 0 {
    switch c {
    case ' ': sign = 1
    case '\t': sign = -1
    case '\r', '\n': error("number error")
    }
  } else {
    switch c {
    case ' ':
      abs <<= 1
    case '\t':
      abs <<= 1
      abs |= 1
    case '\r', '\n':
      num := sign * abs
      debug(fmt.Sprintf("[%d]\n", num))
      sign = 0
      abs = 0
      push_instruction([3]int{current_instruction, 1, num})
      if current_instruction == FLOW_MARK { marks[num] = pc - 1 }
      set_stage(STAGE_IMP)
    }
  }
}

/*
 * Eval
 */
func eval() {
  max_pc := pc
  for pc = 0; pc < max_pc; pc++ {
    //dump_stack()
    instruction := program[pc]
    param := instruction[2]
    switch instruction[0] {
    case STACK_PUSH:
      debug(fmt.Sprintf("(STACK_PUSH:%d)\n", param))
      push_stack(param)
    case STACK_DUPLICATE:
      debug(fmt.Sprintf("(STACK_DUPLICATE)\n"))
      push_stack(peep_stack())
    case STACK_COPY:
      debug(fmt.Sprintf("(STACK_COPY)\n"))
      push_stack(stack[sc - param])
    case STACK_SWAP:
      debug(fmt.Sprintf("(STACK_SWAP)\n"))
      stack[sc-1], stack[sc-2] = stack[sc-2], stack[sc-1]
    case STACK_DISCARD:
      debug(fmt.Sprintf("(STACK_DISCARD)\n"))
      sc--
    case STACK_SLIDE:
      debug(fmt.Sprintf("(STACK_SLIDE)\n"))
      // TODO

    case ARITHMETIC_ADD:
      debug(fmt.Sprintf("(ARITHMETIC_ADD)\n"))
      val1 := pop_stack()
      val2 := pop_stack()
      push_stack(val2 + val1)
    case ARITHMETIC_SUB:
      debug(fmt.Sprintf("(ARITHMETIC_SUB)\n"))
      val1 := pop_stack()
      val2 := pop_stack()
      push_stack(val2 - val1)
    case ARITHMETIC_MUL:
      debug(fmt.Sprintf("(ARITHMETIC_MUL)\n"))
      val1 := pop_stack()
      val2 := pop_stack()
      push_stack(val2 * val1)
    case ARITHMETIC_DIV:
      debug(fmt.Sprintf("(ARITHMETIC_DIV)\n"))
      val1 := pop_stack()
      val2 := pop_stack()
      push_stack(val2 / val1)
    case ARITHMETIC_MOD:
      debug(fmt.Sprintf("(ARITHMETIC_MOD)\n"))
      val1 := pop_stack()
      val2 := pop_stack()
      push_stack(val2 % val1)

    case HEAP_STORE:
      debug(fmt.Sprintf("(HEAP_STORE)\n"))
      val := pop_stack()
      addr := pop_stack()
      heap[addr] = val
    case HEAP_RETRIEVE:
      debug(fmt.Sprintf("(HEAP_RETRIEVE)\n"))
      addr := pop_stack()
      push_stack(heap[addr])

    case FLOW_MARK:
      debug(fmt.Sprintf("(FLOW_MARK)\n"))
      // do nothing because marking is done in parse phase
    case FLOW_CALL_SUB:
      debug(fmt.Sprintf("(FLOW_CALL_SUB:%d)\n", param))
      push_call_stack()
      pc = marks[param]
    case FLOW_JUMP:
      debug(fmt.Sprintf("(FLOW_JUMP:%d)\n", param))
      pc = marks[param]
    case FLOW_JUMP_ZERO:
      debug(fmt.Sprintf("(FLOW_JUMP_ZERO:%d)\n", param))
      if pop_stack() == 0 {
        pc = marks[param]
      }
    case FLOW_JUMP_NEGATIVE:
      debug(fmt.Sprintf("(FLOW_JUMP_NEGATIVE:%d)\n", param))
      if pop_stack() < 0 {
        pc = marks[param]
      }
    case FLOW_END_SUB:
      debug(fmt.Sprintf("(FLOW_END_SUB)\n"))
      pop_call_stack()
    case FLOW_END:
      debug(fmt.Sprintf("(FLOW_END)\n"))
      os.Exit(0)

    case IO_OUT_CHAR:
      debug(fmt.Sprintf("(IO_OUT_CHAR)\n"))
      fmt.Printf("%c", pop_stack())
    case IO_OUT_NUM:
      debug(fmt.Sprintf("(IO_OUT_NUM)\n"))
      fmt.Printf("%d", pop_stack())
    case IO_IN_CHAR:
      debug(fmt.Sprintf("(IO_IN_CHAR)\n"))
      line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
      chr := int(line[0])
      heap[pop_stack()] = chr
    case IO_IN_NUM:
      debug(fmt.Sprintf("(IO_IN_NUM)\n"))
      line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
      num, _ := strconv.Atoi(line[:len(line)-1])
      heap[pop_stack()] = num
    }
  }
}

/*
 * Utilities
 */
func set_stage(stage int) {
  current_stage = stage
  last_char = NULL_CHAR
}

func push_instruction(cmd [3]int) {
  program[pc] = cmd
  pc++
}

func push_stack(val int) {
  stack[sc] = val
  sc++
}

func pop_stack() int {
  val := stack[sc-1]
  sc--
  return val
}

func peep_stack() int {
  return stack[sc-1]
}

func push_call_stack() {
  call_stack[csc] = pc
  csc++
}

func pop_call_stack() int {
  pc = call_stack[csc-1]
  csc--
  return pc
}

func error(message string) {
  fmt.Fprintf(os.Stderr, "%s\n", message)
  os.Exit(1)
}

/*
 * for Debug
 */
func dump() {
  dump_program()
  dump_marks()
  dump_stack()
}

func dump_program() {
  for i := 0; i < pc; i++ {
    instruction := program[i]
    var instruction_type string
    switch instruction[0] {
    case STACK_PUSH: instruction_type = "STACK_PUSH"
    case STACK_DUPLICATE: instruction_type = "STACK_DUPLICATE"
    case STACK_COPY: instruction_type = "STACK_COPY"
    case STACK_SWAP: instruction_type = "STACK_SWAP"
    case STACK_DISCARD: instruction_type = "STACK_DISCARD"
    case STACK_SLIDE: instruction_type = "STACK_SLIDE"
    case ARITHMETIC_ADD: instruction_type = "ARITHMETIC_ADD"
    case ARITHMETIC_SUB: instruction_type = "ARITHMETIC_SUB"
    case ARITHMETIC_MUL: instruction_type = "ARITHMETIC_MUL"
    case ARITHMETIC_DIV: instruction_type = "ARITHMETIC_DIV"
    case ARITHMETIC_MOD: instruction_type = "ARITHMETIC_MOD"
    case HEAP_STORE: instruction_type = "HEAP_STORE"
    case HEAP_RETRIEVE: instruction_type = "HEAP_RETRIEVE"
    case FLOW_MARK: instruction_type = "FLOW_MARK"
    case FLOW_CALL_SUB: instruction_type = "FLOW_CALL_SUB"
    case FLOW_JUMP: instruction_type = "FLOW_JUMP"
    case FLOW_JUMP_ZERO: instruction_type = "FLOW_JUMP_ZERO"
    case FLOW_JUMP_NEGATIVE: instruction_type = "FLOW_JUMP_NEGATIVE"
    case FLOW_END_SUB: instruction_type = "FLOW_END_SUB"
    case FLOW_END: instruction_type = "FLOW_END"
    case IO_OUT_CHAR: instruction_type = "IO_OUT_CHAR"
    case IO_OUT_NUM: instruction_type = "IO_OUT_NUM"
    case IO_IN_CHAR: instruction_type = "IO_IN_CHAR"
    case IO_IN_NUM: instruction_type = "IO_IN_NUM"
    }
    param := ""
    if instruction[1] == 1 {
      param = fmt.Sprintf("%d", instruction[2])
    }
    fmt.Printf("[%s] %s\n", instruction_type, param)
  }
}

func dump_marks() {
  for k, v := range marks {
    fmt.Printf("%d -> %d\n", k, v)
  }
}

func dump_stack() {
  fmt.Printf("stack:(%d) [", sc)
  for i := 0; i < sc; i++ {
    fmt.Printf("%d ", stack[sc])
  }
  fmt.Printf("]\n")
}

func debug(message string) {
  if is_debug { fmt.Printf("%s", message) }
}

/*
 * Main
 */
func main() {
  flag.BoolVar(&is_debug, "d", false, "debug mode")
  flag.Parse()
  if flag.NArg() == 0 {
    fmt.Printf("filename required")
    return
  }

  filename := flag.Arg(0)
  file, err := os.Open(filename, os.O_RDONLY, 0666)
  if (err == nil) {
    parse(file)
    eval()
  }
}
