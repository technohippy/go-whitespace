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
  STAGE_COMMAND
  STAGE_NUMBER
  STAGE_LABEL

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

// Flow
var (
  last_char = NULL_CHAR
  current_stage = STAGE_IMP
  current_imp int
  current_instruction int
)

var program [STACK_SIZE][3]int
var pc = 0

var marks = make(map[string] int)

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
        case STAGE_COMMAND: parse_instruction(c)
        case STAGE_NUMBER: parse_number(c)
        case STAGE_LABEL: parse_label(c)
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
      set_stage(STAGE_COMMAND)
    case '\t':
      // [Tab][Space]
      debug("[IMP:ARITHMETIC] ")
      current_imp = IMP_ARITHMETIC
      set_stage(STAGE_COMMAND)
    }
  case '\t':
    switch last_char {
    case '\t':
      // [Tab][Tab]
      debug("[IMP:HEAP] ")
      current_imp = IMP_HEAP
      set_stage(STAGE_COMMAND)
    }
  case '\r', '\n':
    switch last_char {
    case NULL_CHAR:
      // [LF]
      debug("[IMP:FLOW] ")
      current_imp = IMP_FLOW
      set_stage(STAGE_COMMAND)
    case '\t':
      // [Tab][LF]
      debug("[IMP:I/O] ")
      current_imp = IMP_IO
      set_stage(STAGE_COMMAND)
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
    parse_command_stack(c)
  case IMP_ARITHMETIC:
    parse_command_arithmetic(c)
  case IMP_HEAP:
    parse_command_heap(c)
  case IMP_FLOW:
    parse_command_flow(c)
  case IMP_IO:
    parse_command_io(c)
  }
  if current_stage == STAGE_COMMAND {
    last_char = c
  }
}

// Stack Manipulation (IMP: [Space])
func parse_command_stack(c int) {
  switch c {
  case ' ':
    switch last_char {
    case NULL_CHAR:
      // [Space] Number : Push the number onto the stack
      debug("[Push] ")
      current_instruction = STACK_PUSH
      set_stage(STAGE_NUMBER)
    case '\r', '\n':
      // [LF][Space] : Duplicate the top item on the stack
      debug("[Duplicate]\n")
      push_instruction([3]int{STACK_DUPLICATE, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Space] Number : Copy the nth item on the stack onto the top of the stack
      debug("[Copy] ")
      current_instruction = STACK_COPY
      set_stage(STAGE_NUMBER)
    }
  case '\t':
    switch last_char {
    case '\r', '\n':
      // [LF][Tab] : Swap the top two items on the stack
      debug("[Swap]\n")
      push_instruction([3]int{STACK_SWAP, 0, 0})
      set_stage(STAGE_IMP)
    default:
      error(fmt.Sprintf("Parse Error: stack: '%c'", c))
    }
  case '\r', '\n':
    switch last_char {
    case NULL_CHAR:
      // do nothing
    case '\r', '\n':
      // [LF][LF] : Discard the top item on the stack
      debug("[Discard]\n")
      push_instruction([3]int{STACK_DISCARD, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][LF] Number : Slide n items off the stack, keeping the top item
      debug("[Slide] ")
      current_instruction = STACK_SLIDE
      set_stage(STAGE_NUMBER)
    default:
      error(fmt.Sprintf("Parse Error: stack: '%c'", c))
    }
  }
}

// Arithmetic (IMP: [Tab][Space])
func parse_command_arithmetic(c int) {
  switch c {
  case ' ':
    switch last_char {
    case ' ':
      // [Space][Space] : Addition
      debug("[Add]\n")
      push_instruction([3]int{ARITHMETIC_ADD, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Space] : Integer Division
      debug("[Devision]\n")
      push_instruction([3]int{ARITHMETIC_DIV, 0, 0})
      set_stage(STAGE_IMP)
    }
  case '\t':
    switch last_char {
    case ' ':
      // [Space][Tab] : Subtraction
      debug("[Subtraction]\n")
      push_instruction([3]int{ARITHMETIC_SUB, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Tab] : Modulo
      debug("[Modulo]\n")
      push_instruction([3]int{ARITHMETIC_MOD, 0, 0})
      set_stage(STAGE_IMP)
    }
  case '\r', '\n':
    switch last_char {
    case ' ':
      // [Space][LF] : Multiplication
      debug("[Multiplication]\n")
      push_instruction([3]int{ARITHMETIC_MUL, 0, 0})
      set_stage(STAGE_IMP)
    default:
      error(fmt.Sprintf("Parse Error: arithmetic: '%c'", c))
    }
  }
}

// Heap Access (IMP: [Tab][Tab])
func parse_command_heap(c int) {
  switch c {
  case ' ':
    // [Space] : Store
    debug("[Store]\n")
    push_instruction([3]int{HEAP_STORE, 0, 0})
    set_stage(STAGE_IMP)
  case '\t':
    // [Tab] : Retrieve
    debug("[Retrieve]\n")
    push_instruction([3]int{HEAP_RETRIEVE, 0, 0})
    set_stage(STAGE_IMP)
  default:
    error(fmt.Sprintf("Parse Error: heap: '%c'", c))
  }
}

// Flow Control (IMP: [LF])
func parse_command_flow(c int) {
  switch c {
  case ' ':
    switch last_char {
    case ' ':
      // [Space][Space] Label : Mark a location in the program
      debug("[*MARK*] ")
      current_instruction = FLOW_MARK
      set_stage(STAGE_LABEL)
    case '\t':
      // [Tab][Space] Label : Jump to a label if the top of the stack is zero
      debug("[Jump Zero] ")
      current_instruction = FLOW_JUMP_ZERO
      set_stage(STAGE_LABEL)
    }
  case '\t':
    switch last_char {
    case ' ':
      // [Space][Tab] Label : Call a subroutine
      debug("[Call] ")
      current_instruction = FLOW_CALL_SUB
      set_stage(STAGE_LABEL)
    case '\t':
      // [Tab][Tab] Label : Jump to a label if the top of the stack is negative
      debug("[Jump Neg] ")
      current_instruction = FLOW_JUMP_NEGATIVE
      set_stage(STAGE_LABEL)
    }
  case '\r', '\n':
    switch last_char {
    case ' ':
      // [Space][LF] Label : Jump unconditionally to a label
      debug("[Jump] ")
      current_instruction = FLOW_JUMP
      set_stage(STAGE_LABEL)
    case '\t':
      // [Tab][LF] : End a subroutine and transfer control back to the caller
      debug("[End Sub]\n")
      push_instruction([3]int{FLOW_END_SUB, 0, 0})
      set_stage(STAGE_IMP)
    case '\r', '\n':
      // [LF][LF] : End the program
      debug("[End]\n")
      push_instruction([3]int{FLOW_END, 0, 0})
      set_stage(STAGE_IMP)
    }
  }
}

// I/O (IMP: [Tab][LF])
func parse_command_io(c int) {
  switch c {
  case ' ':
    switch last_char {
    case ' ':
      // [Space][Space] : Output the character at the top of the stack
      debug("[Out Char]\n")
      push_instruction([3]int{IO_OUT_CHAR, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Space] : Read a character and place it in the location given by the top of the stack
      debug("[In Char]\n")
      push_instruction([3]int{IO_IN_CHAR, 0, 0})
      set_stage(STAGE_IMP)
    }
  case '\t':
    switch last_char {
    case ' ':
      // [Space][Tab] : Output the number at the top of the stack
      debug("[Out Num]\n")
      push_instruction([3]int{IO_OUT_NUM, 0, 0})
      set_stage(STAGE_IMP)
    case '\t':
      // [Tab][Tab] : Read a number and place it in the location given by the top of the stack
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

// Number
func parse_number(c int) {
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
      push_instruction([3]int{current_instruction, sign, num})
      //if current_instruction == FLOW_MARK { marks[fmt.Sprintf("%d/%d", sign, num)] = pc - 1 }
      set_stage(STAGE_IMP)
      sign = 0
      abs = 0
    }
  }
}

var label = 1
// label
func parse_label(c int) {
  switch c {
  case ' ':
    label <<= 1
  case '\t':
    label <<= 1
    label |= 1
  case '\r', '\n':
    debug(fmt.Sprintf("[%d]\n", label))
    push_instruction([3]int{current_instruction, 0, label})
    if current_instruction == FLOW_MARK { marks[fmt.Sprintf("%d", label)] = pc - 1 }
    set_stage(STAGE_IMP)
    label = 1
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
    command := instruction[0]
    param := instruction[2]
    switch command {
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
      pc = marks[fmt.Sprintf("%d", param)]
    case FLOW_JUMP:
      debug(fmt.Sprintf("(FLOW_JUMP:%d)\n", param))
      pc = marks[fmt.Sprintf("%d", param)]
    case FLOW_JUMP_ZERO:
      debug(fmt.Sprintf("(FLOW_JUMP_ZERO:%d)\n", param))
      if pop_stack() == 0 {
        pc = marks[fmt.Sprintf("%d", param)]
      }
    case FLOW_JUMP_NEGATIVE:
      debug(fmt.Sprintf("(FLOW_JUMP_NEGATIVE:%d)\n", param))
      if pop_stack() < 0 {
        pc = marks[fmt.Sprintf("%d", param)]
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
      num, _ := strconv.Atoi(line[0:len(line)-1])
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
