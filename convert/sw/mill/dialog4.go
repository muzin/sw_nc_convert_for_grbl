package sw_mill

import (
	"fmt"
	"github.com/muzin/go_rt/collection/array_list"
	"github.com/muzin/go_rt/collection/hash_map"
	rt_str "github.com/muzin/go_rt/lang/str"
	"github.com/muzin/go_rt/try"
	"strings"
	"sw_nc_convert_for_grbl/cnc_gcode"
)

var ConvertException = try.DeclareException("ConvertException")

type Dialog4MillConverter struct {
	CncGcodeFile *cnc_gcode.CncGcodeFile
}

func (this *Dialog4MillConverter) Convert(cncGcodeFile *cnc_gcode.CncGcodeFile) *cnc_gcode.CncGcodeFile {

	if cncGcodeFile == nil {
		try.Throw(ConvertException.NewThrow("CncGcodeFile is nil"))
	}

	this.CncGcodeFile = cncGcodeFile

	// 转换 手动改换刀 指令
	// 转换
	//		G17 T0*
	// 为
	// 		M05
	//		M00
	//      G04 P5
	this.covertToolChanging(cncGcodeFile)

	// 刷新行号
	cncGcodeFile.GetHead().FlushLineNumber(1)

	// 转换 指针行号
	this.ConvertLineNumberPointer()

	// 转换 重复Gcode
	this.ConvertRepeatedGcodeWord()

	return cncGcodeFile
}

func (this *Dialog4MillConverter) covertToolChanging(cncGcodeFile *cnc_gcode.CncGcodeFile) {

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

//type GcodeSlice struct {
//	Line  int
//	Codes []string
//}

// 查找 换刀 指令的索引
func (this *Dialog4MillConverter) findToolChangingGcode() *array_list.ArrayList {

	cncGcodeFile := this.CncGcodeFile

	val := array_list.NewArrayList()

	gcodeCommand := cncGcodeFile.GetHead()

	for ; gcodeCommand != nil; gcodeCommand = gcodeCommand.Next {
		cncGcodeCommand := gcodeCommand

		if cncGcodeCommand.LineNumber > 0 {

			codes := cncGcodeCommand.Codes

			existsG17 := false
			existsT0 := false

			for j := 0; j < codes.Size(); j++ {
				subWord := codes.Get(j).(*cnc_gcode.CncWord)

				if !existsG17 {
					existsG17 = strings.HasPrefix(subWord.String(), "G17")
				}

				if !existsT0 {
					existsT0 = strings.HasPrefix(subWord.String(), "T")
				}

			}

			if existsG17 && existsT0 {
				val.Add(gcodeCommand.LineNumber)
			}

		}

	}

	return val
}

func (this *Dialog4MillConverter) ConvertLineNumberPointer() {

	cncGcodeFile := this.CncGcodeFile

	lineNumberPointerGcodeMap := this.FindLineNumberPointerGcodeMap()
	lineNumberPointerCommandMap := this.FindLineNumberPointerCommandMap()

	keys := lineNumberPointerCommandMap.Keys()
	for i := 0; i < len(keys); i++ {
		index := keys[i].(int)
		currentGcodeCommand := cncGcodeFile.GetByLineNumber(index)
		codeString := currentGcodeCommand.GetCodeString()

		if lineNumberPointerGcodeMap.ContainsKey(codeString) {
			cncGcodeCommand := lineNumberPointerGcodeMap.Get(codeString).(*cnc_gcode.CncGcodeCommand)
			cncGcodeString := cncGcodeCommand.GetCodeString()

			currentGcodeCommand.Codes.Clear()
			currentGcodeCommand.Codes = currentGcodeCommand.ParseCncWord(cncGcodeString)
		}
	}

}

func (this *Dialog4MillConverter) FindLineNumberPointerGcodeMap() *hash_map.HashMap {

	cncGcodeFile := this.CncGcodeFile

	gcodeCommand := cncGcodeFile.GetHead()

	hashMap := hash_map.NewHashMap()

	for ; gcodeCommand != nil; gcodeCommand = gcodeCommand.Next {
		if strings.HasPrefix(gcodeCommand.LineNumberStr, "*") {
			hashMap.Put("N"+gcodeCommand.LineNumberStr, gcodeCommand)
		}
	}

	return hashMap
}

func (this *Dialog4MillConverter) FindLineNumberPointerCommandMap() *hash_map.HashMap {

	cncGcodeFile := this.CncGcodeFile

	gcodeCommand := cncGcodeFile.GetHead()

	hashMap := hash_map.NewHashMap()

	for ; gcodeCommand != nil; gcodeCommand = gcodeCommand.Next {
		if strings.HasPrefix(gcodeCommand.GetCodeString(), "N*") {
			hashMap.Put(gcodeCommand.LineNumber, gcodeCommand)
		}
	}

	return hashMap
}

// 转换重复Gcode
func (this *Dialog4MillConverter) ConvertRepeatedGcodeWord() {
	//repeatedGcodeWordCommands := this.FindRepeatedGcodeWordCommand()

	cncGcodeFile := this.CncGcodeFile

	// TODO 拆分 后面的  G04 F10
	g81GcodeMap := cncGcodeFile.FindExistGcodeWordCommand("G81")

	g81GcodeCommandPtrList := g81GcodeMap.Values()
	for i := 0; i < len(g81GcodeCommandPtrList); i++ {
		cncGcodeCommandItem := g81GcodeCommandPtrList[i].(*cnc_gcode.CncGcodeCommand)

		// 移除掉 G04  F10
		cncGcodeCommandItem.PopCncWord()
		cncGcodeCommandItem.PopCncWord()

		// 添加 R指令
		cncGcodeCommandItem.AddCncWord(cnc_gcode.NewCncWord("R0"))

	}

	// TODO 验证 是否需要 合并  下面 的 坐标移动
	g82GcodeMap := cncGcodeFile.FindExistGcodeWordCommand("G82")

	g82GcodeMap.Get(0)

	// TODO 拆分 后面 重复 Z
	g83GcodeMap := cncGcodeFile.FindExistGcodeWordCommand("G83")

	g83GcodeCommandPtrList := g83GcodeMap.Values()
	for i := 0; i < len(g83GcodeCommandPtrList); i++ {
		cncGcodeCommandItem := g83GcodeCommandPtrList[i].(*cnc_gcode.CncGcodeCommand)

		var zValArr []float64
		var firstZVal float64
		var rVal float64 //在绝对方式下指定Z轴方向R点的位置,增量方式下指定从初始点到R点的距离。
		var fVal int     //进给速度
		var qVal int     //进刀量
		firstZVal = 0

		if cncGcodeCommandItem.GetCncWordLength() <= 1 {
			continue
		}

		lastCncWordF := cncGcodeCommandItem.GetLastCncWordByWord("F")
		fVal = int(lastCncWordF.GetValOfNumber())

		// 获取所有z
		lastCncWordZ := cncGcodeCommandItem.RemoveLastCncWordByWord("Z")
		for ; lastCncWordZ != nil; lastCncWordZ = cncGcodeCommandItem.RemoveLastCncWordByWord("Z") {
			valOfNumber := lastCncWordZ.GetValOfNumber()
			zValArr = append(zValArr, valOfNumber)
		}

		firstZVal = zValArr[0]
		rVal = firstZVal
		qVal = fVal / (len(zValArr) * 2)

		cncGcodeCommandStr := strings.TrimSpace(cncGcodeCommandItem.String())

		cncGcodeCommandItem.AddCncWord(cnc_gcode.NewCncWord("Z" + rt_str.Strval(firstZVal)))
		cncGcodeCommandItem.AddCncWord(cnc_gcode.NewCncWord("R" + rt_str.Strval(rVal)))
		cncGcodeCommandItem.AddCncWord(cnc_gcode.NewCncWord("Q" + rt_str.Strval(qVal)))

		for j := len(zValArr) - 1; j > 0; j-- {
			zValItem := zValArr[j]
			cncGcodeCommandItem.AddToBack(
				cncGcodeFile.ParseToCncGcodeCommand(fmt.Sprintf("%s Z%.3f R%s Q%d",
					cncGcodeCommandStr,
					zValItem,
					rt_str.Strval(rVal),
					qVal,
				)))
		}
	}
}
