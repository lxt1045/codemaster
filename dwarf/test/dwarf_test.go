package test

import (
	"debug/dwarf"
	"debug/elf"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hitzhangjie/codemaster/dwarf/frame"
	"github.com/hitzhangjie/codemaster/dwarf/godwarf"
	"github.com/hitzhangjie/codemaster/dwarf/line"
	"github.com/hitzhangjie/codemaster/dwarf/reader"
	"github.com/hitzhangjie/codemaster/dwarf/regnum"
)

func Test_ElfReadDWARF(t *testing.T) {
	f, err := elf.Open("fixtures/elf_read_dwarf")
	// f, err := elf.Open("fixtures/fixtures.exe")
	assert.Nil(t, err)

	sections := []string{
		"abbrev",
		"line",
		"frame",
		"pubnames",
		"pubtypes",
		//"gdb_script",
		"info",
		"loc",
		"ranges",
	}

	for _, s := range sections {
		b, err := godwarf.GetDebugSection(f, s)
		assert.Nil(t, err)
		t.Logf(".[z]debug_%s data size: %d", s, len(b))
	}
}

func Test_PeReadDWARF(t *testing.T) {
	f, err := pe.Open("fixtures/fixtures.exe")
	assert.Nil(t, err)

	sections := []string{
		"abbrev",
		"line",
		"frame",
		"pubnames",
		"pubtypes",
		//"gdb_script",
		"info",
		"loc",
		"ranges",
	}

	for _, s := range sections {
		b, err := godwarf.GetDebugSectionPE(f, s)
		assert.Nil(t, err)
		t.Logf(".[z]debug_%s data size: %d", s, len(b))
	}
}
func Test_DWARFReadTypes(t *testing.T) {
	// f, err := elf.Open("fixtures/elf_read_dwarf")
	f, err := pe.Open("D:/project/go/src/github.com/lxt1045/errors/cmd/cmd.exe") // 编译时去掉DWARF: go build --ldflags="-w -s"
	assert.Nil(t, err)

	dat, err := f.DWARF()
	assert.Nil(t, err)

	rd := reader.New(dat)

	for {
		e, err := rd.NextType()
		if err != nil {
			break
		}
		if e == nil {
			break
		}
		t.Logf("read type: %s", e.Val(dwarf.AttrName))
	}
}

func Test_DWARFReadTypes2(t *testing.T) {
	// f, err := elf.Open("fixtures/elf_read_dwarf")
	f, err := pe.Open("fixtures/fixtures.exe")
	assert.Nil(t, err)

	dat, err := f.DWARF()
	assert.Nil(t, err)

	var cuName string
	var rd = reader.New(dat)
	for {
		entry, err := rd.Next()
		if err != nil {
			break
		}
		if entry == nil {
			break
		}

		switch entry.Tag {
		case dwarf.TagCompileUnit:
			cuName = entry.Val(dwarf.AttrName).(string)
			t.Logf("- CompilationUnit[%s]", cuName)
		case dwarf.TagArrayType,
			dwarf.TagBaseType,
			dwarf.TagClassType,
			dwarf.TagStructType,
			dwarf.TagUnionType,
			dwarf.TagConstType,
			dwarf.TagVolatileType,
			dwarf.TagRestrictType,
			dwarf.TagEnumerationType,
			dwarf.TagPointerType,
			dwarf.TagSubroutineType,
			dwarf.TagTypedef,
			dwarf.TagUnspecifiedType:
			t.Logf("  cu[%s] define [%s]", cuName, entry.Val(dwarf.AttrName))
		}
	}
}

func Test_DWARFReadTypes3(t *testing.T) {
	f, err := elf.Open("fixtures/elf_read_dwarf")
	assert.Nil(t, err)

	dat, err := f.DWARF()
	assert.Nil(t, err)

	var rd = reader.New(dat)

	entry, err := rd.SeekToTypeNamed("main.Student")
	assert.Nil(t, err)
	fmt.Println(entry)
}

func Test_DWARFReadVariable(t *testing.T) {
	f, err := elf.Open("fixtures/elf_read_dwarf")
	assert.Nil(t, err)

	dat, err := f.DWARF()
	assert.Nil(t, err)

	var rd = reader.New(dat)
	for {
		entry, err := rd.Next()
		if err != nil {
			break
		}
		if entry == nil {
			break
		}
		// 只查看变量
		if entry.Tag != dwarf.TagVariable {
			continue
		}
		// 只查看变量名为s的变量
		if entry.Val(dwarf.AttrName) != "s" {
			continue
		}
		// 通过offset限制，只查看main.main中定义的变量名为s的变量
		// 这里的0x432b9是结合`objdump --dwarf=info` 中的结果来硬编码的
		if entry.Val(dwarf.AttrType).(dwarf.Offset) != dwarf.Offset(0x432b9) {
			continue
		}

		// 查看变量s的DIE
		fmt.Println("found the variable[s]")
		fmt.Println("DIE variable:", entry)

		// 查看变量s对应的类型的DIE
		variableTypeEntry, err := rd.SeekToType(entry, true, true)
		assert.Nil(t, err)
		fmt.Println("DIE type:", variableTypeEntry)

		// 查看变量s对应的地址 [lowpc, highpc, instruction]
		fmt.Println("location:", entry.Val(dwarf.AttrLocation))

		// 最后在手动校验下main.Student的类型与上面看到的变量的类型是否一致
		// 应该满足：main.Student DIE的位置 == 变量的类型的位置偏移量
		typeEntry, err := rd.SeekToTypeNamed("main.Student")
		assert.Nil(t, err)
		assert.Equal(t, typeEntry.Val(dwarf.AttrType), variableTypeEntry.Offset)
		break
	}
}

func Test_DWARFReadFunc(t *testing.T) {
	f, err := elf.Open("fixtures/elf_read_dwarf")
	assert.Nil(t, err)

	dat, err := f.DWARF()
	assert.Nil(t, err)

	rd := reader.New(dat)
	i := 0
	for {
		die, err := rd.Next()
		if err != nil {
			break
		}
		if die == nil {
			break
		}
		if die.Tag == dwarf.TagSubprogram {
			fmt.Println(die)
		}
		i++
	}
	t.Log("i:", i)
}

func Test_DWARFReadLineNoTable(t *testing.T) {
	// f, err := elf.Open("fixtures/elf_read_dwarf")
	f, err := pe.Open("fixtures/fixtures.exe")
	assert.Nil(t, err)

	// dat, err := godwarf.GetDebugSection(f, "line")
	dat, err := godwarf.GetDebugSectionPE(f, "line")
	assert.Nil(t, err)

	lineToPCs := map[int][]uint64{10: nil, 12: nil, 13: nil, 14: nil, 15: nil}

	debuglines := line.ParseAll(dat, nil, nil, 0, true, 8)
	fmt.Println(len(debuglines))
	for _, line := range debuglines {
		// if len(line.FileNames) > 0 {
		// 	fmt.Printf("idx-%d\tinst:%v\n", i, line.FileNames[0].Path)
		// }
		// line.AllPCsForFileLines("/root/dwarftest/dwarf/test/fixtures/elf_read_dwarf.go", lineToPCs)
		line.AllPCsForFileLines("D:/project/go/src/github.com/hitzhangjie/codemaster/dwarf/test/fixtures/elf_read_dwarf.go", lineToPCs)
	}

	for line, pcs := range lineToPCs {
		fmt.Printf("lineNo:[elf_read_dwarf.go:%d] -> PC:%#x\n", line, pcs)
	}
}

func Test_DWARFReadCFITable(t *testing.T) {
	f, err := elf.Open("fixtures/elf_read_dwarf")
	assert.Nil(t, err)

	// 解析.[z]debug_frame中CFI信息表
	dat, err := godwarf.GetDebugSection(f, "frame")
	assert.Nil(t, err)
	fdes, err := frame.Parse(dat, binary.LittleEndian, 0, 8, 0)
	assert.Nil(t, err)
	assert.NotEmpty(t, fdes)

	//for idx, fde := range fdes {
	//	fmt.Printf("fde[%d], begin:%#x, end:%#x\n", idx, fde.Begin(), fde.End())
	//}

	for _, fde := range fdes {
		if !fde.Cover(0x4b8640) {
			continue
		}
		fmt.Printf("address 0x4b8640 is covered in FDE[%#x,%#x]\n", fde.Begin(), fde.End())
		fc := fde.EstablishFrame(0x4b8640)
		fmt.Printf("retAddReg: %s\n", regnum.AMD64ToName(fc.RetAddrReg))
		switch fc.CFA.Rule {
		case frame.RuleCFA:
			fmt.Printf("cfa: rule:RuleCFA, CFA=(%s)+%#x\n", regnum.ARM64ToName(fc.CFA.Reg), fc.CFA.Offset)
		default:
		}
	}
}
