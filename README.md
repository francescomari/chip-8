# CHIP-8

An emulator for the CHIP-8 virtual machine.

## Usage

Open a rom with the `chip8` program.

```sh
go run ./cmd/chip8 roms/7-beep.ch8
```

Only CHIP-8 instructions are supported. Invalid roms will trigger a panic in the
emulator.

## Debugger

While running a rom, you can toggle debug mode by pressing the `P` key. This
will open a debug panel that allows you to step over the rom one instruction at
a time, or simulate the passage of time. You can run the rom normally by
pressing `P` again.

If you want to start the emulator in debug mode, add the `-debug` flag to the
command line:

```sh
go run ./cmd/chip8 -debug roms/7-beep.ch8
```

## References

- [CHIP-8 on Wikipedia](https://en.wikipedia.org/wiki/CHIP-8)
- [How to Write an Emulator (CHIP-8 interpreter)](https://multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter/)
- [Cowgod's CHIP-8 Technical Reference](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM)
- [CHIP-8 Research Facility](https://chip-8.github.io/)
- [CHIP-8 Test Suite](https://github.com/Timendus/chip8-test-suite)
