package convert

import (
	"sw_nc_convert_for_grbl/cnc_gcode"
	sw_mill "sw_nc_convert_for_grbl/convert/sw/mill"
)

type CncConverter interface {
	Convert(file *cnc_gcode.CncGcodeFile) *cnc_gcode.CncGcodeFile
}

func GetCncConverter(name string) CncConverter {
	if "sw/mill/dialog4" == name {
		return &sw_mill.Dialog4MillConverter{}
	} else if "sw/mill/m3axis" == name {
		return &sw_mill.M3AxisMillConverter{}
	} else {
		return nil
	}
}
