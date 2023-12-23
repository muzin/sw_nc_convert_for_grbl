package sw_mill

import (
	"github.com/muzin/go_rt/collection/array_list"
	"github.com/muzin/go_rt/try"
	"strconv"
	"strings"
	"sw_nc_convert_for_grbl/cnc_gcode"
)

type M3AxisMillConverter struct {
	CncGcodeFile *cnc_gcode.CncGcodeFile
}

func (this *M3AxisMillConverter) Convert(cncGcodeFile *cnc_gcode.CncGcodeFile) *cnc_gcode.CncGcodeFile {

	if cncGcodeFile == nil {
		try.Throw(ConvertException.NewThrow("CncGcodeFile is nil"))
	}

	this.CncGcodeFile = cncGcodeFile

	// 转换 手动改换刀 指令
	// 转换
	//		T0* M06
	// 为
	// 		M05
	//		M00
	//      (T0* M06)
	//      G04 P5
	this.covertToolChanging(cncGcodeFile)

	// 刷新行号
	cncGcodeFile.GetHead().FlushLineNumber(1)

	// 转换 G82 Gcode
	this.ConvertG82Gcode()

	return cncGcodeFile
}

func (this *M3AxisMillConverter) covertToolChanging(cncGcodeFile *cnc_gcode.CncGcodeFile) {

	// 获取 换刀 指令的索引
	// 寻找 G17 T0* 相关的指令，在其前后加入相关指令
	toolChangingGcodeIndex := this.findToolChangingGcode()

	for i := 0; i < toolChangingGcodeIndex.Size(); i++ {

		// 第一次换刀不做处理
		if i == 0 {
			continue
		}

		currentCncGcodeCommand := cncGcodeFile.GetByLineNumber(toolChangingGcodeIndex.Get(i).(int))

		// 在换刀前添加 关闭主轴/程序暂停指令   M05: 关闭主轴  M00: 程序暂停
		currentCncGcodeCommand.AddToFront(cncGcodeFile.ParseToCncGcodeCommand("N000 M05"))
		currentCncGcodeCommand.AddToFront(cncGcodeFile.ParseToCncGcodeCommand("N000 M00"))

		nextCncGcodeCommand := currentCncGcodeCommand.Next
		if !nextCncGcodeCommand.ExistsCncWord("M03") {
			nextCncGcodeCommand.AddCncWord(cnc_gcode.NewCncWord("M03"))
		}

		// 主轴开启后延迟5秒种  G04 P5 : 延迟5秒种
		nextCncGcodeCommand.AddToBack(cncGcodeFile.ParseToCncGcodeCommand("N000 G04 P5"))

	}

}

// 查找 换刀 指令的索引
func (this *M3AxisMillConverter) findToolChangingGcode() *array_list.ArrayList {

	cncGcodeFile := this.CncGcodeFile

	val := array_list.NewArrayList()

	gcodeCommand := cncGcodeFile.GetHead()

	for ; gcodeCommand != nil; gcodeCommand = gcodeCommand.Next {
		cncGcodeCommand := gcodeCommand

		if cncGcodeCommand.LineNumber > 0 {

			codes := cncGcodeCommand.Codes

			existsG06 := false
			existsT0 := false

			for j := 0; j < codes.Size(); j++ {
				subWord := codes.Get(j).(*cnc_gcode.CncWord)

				if !existsG06 {
					existsG06 = strings.HasPrefix(subWord.String(), "M06")
				}

				if !existsT0 {
					existsT0 = strings.HasPrefix(subWord.String(), "T0")
				}

			}

			if existsG06 && existsT0 {
				val.Add(gcodeCommand.LineNumber)
			}

		}

	}

	return val
}

// 转换G82 Gcode
// 将P指令延迟的毫秒转化为秒
func (this *M3AxisMillConverter) ConvertG82Gcode() {
	//repeatedGcodeWordCommands := this.FindRepeatedGcodeWordCommand()

	cncGcodeFile := this.CncGcodeFile

	// TODO 拆分 后面的  G04 F10
	g82GcodeMap := cncGcodeFile.FindExistGcodeWordCommand("G82")

	g82GcodeCommandPtrList := g82GcodeMap.Values()
	for i := 0; i < len(g82GcodeCommandPtrList); i++ {
		cncGcodeCommandItem := g82GcodeCommandPtrList[i].(*cnc_gcode.CncGcodeCommand)

		// 查询P指令
		cncWord := cncGcodeCommandItem.GetLastCncWordByWord("P")
		valStr := cncWord.GetValStr()
		val, _ := strconv.Atoi(valStr)
		if val > 0 {
			// sw 后处理 延时 为毫秒 ，转换为grbl的秒
			val /= 1000
		}
		cncWord.Val = strconv.Itoa(val)
	}

}
