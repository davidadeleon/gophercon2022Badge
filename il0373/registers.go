package il0373

/*
   EPD_WHITE = 0b00;
   EPD_BLACK = 0b11;
   EPD_RED = 0b01;
   EPD_GRAY = 0b10;
   EPD_LIGHT = 0b01;
   EPD_DARK = 0b10;
*/

const (
	IL0373_PANEL_SETTING      = 0x00
	IL0373_POWER_SETTING      = 0x01
	IL0373_POWER_OFF          = 0x02
	IL0373_POWER_OFF_SEQUENCE = 0x03
	IL0373_POWER_ON           = 0x04
	IL0373_POWER_ON_MEASURE   = 0x05
	IL0373_BOOSTER_SOFT_START = 0x06
	IL0373_DEEP_SLEEP         = 0x07
	IL0373_DTM1               = 0x10
	IL0373_DATA_STOP          = 0x11
	IL0373_DISPLAY_REFRESH    = 0x12
	IL0373_DTM2               = 0x13
	IL0373_PDTM1              = 0x14
	IL0373_PDTM2              = 0x15
	IL0373_PDRF               = 0x16
	IL0373_LUT1               = 0x20
	IL0373_LUTWW              = 0x21
	IL0373_LUTBW              = 0x22
	IL0373_LUTWB              = 0x23
	IL0373_LUTBB              = 0x24
	IL0373_PLL                = 0x30
	IL0373_CDI                = 0x50
	IL0373_RESOLUTION         = 0x61
	IL0373_VCM_DC_SETTING     = 0x82
)
