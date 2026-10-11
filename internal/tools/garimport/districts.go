package garimport

const (
	fdCentral       = "Центральный"
	fdNorthWest     = "Северо-Западный"
	fdSouth         = "Южный"
	fdNorthCaucasus = "Северо-Кавказский"
	fdVolga         = "Приволжский"
	fdUral          = "Уральский"
	fdSiberia       = "Сибирский"
	fdFarEast       = "Дальневосточный"
)

// federalDistricts maps a GAR region folder code to its federal district,
// named as in hflabs/city (without the "ФО" suffix). Folders 80–82, 84, 85,
// 88 are abolished subjects and are empty; 99 (Байконур) has no district and
// is skipped.
var federalDistricts = map[string]string{
	"01": fdSouth, "02": fdVolga, "03": fdFarEast, "04": fdSiberia,
	"05": fdNorthCaucasus, "06": fdNorthCaucasus, "07": fdNorthCaucasus,
	"08": fdSouth, "09": fdNorthCaucasus, "10": fdNorthWest,
	"11": fdNorthWest, "12": fdVolga, "13": fdVolga,
	"14": fdFarEast, "15": fdNorthCaucasus, "16": fdVolga,
	"17": fdSiberia, "18": fdVolga, "19": fdSiberia, "20": fdNorthCaucasus,
	"21": fdVolga, "22": fdSiberia, "23": fdSouth, "24": fdSiberia,
	"25": fdFarEast, "26": fdNorthCaucasus, "27": fdFarEast,
	"28": fdFarEast, "29": fdNorthWest, "30": fdSouth, "31": fdCentral,
	"32": fdCentral, "33": fdCentral, "34": fdSouth, "35": fdNorthWest,
	"36": fdCentral, "37": fdCentral, "38": fdSiberia, "39": fdNorthWest,
	"40": fdCentral, "41": fdFarEast, "42": fdSiberia, "43": fdVolga,
	"44": fdCentral, "45": fdUral, "46": fdCentral, "47": fdNorthWest,
	"48": fdCentral, "49": fdFarEast, "50": fdCentral,
	"51": fdNorthWest, "52": fdVolga, "53": fdNorthWest,
	"54": fdSiberia, "55": fdSiberia, "56": fdVolga, "57": fdCentral,
	"58": fdVolga, "59": fdVolga, "60": fdNorthWest, "61": fdSouth,
	"62": fdCentral, "63": fdVolga, "64": fdVolga, "65": fdFarEast,
	"66": fdUral, "67": fdCentral, "68": fdCentral, "69": fdCentral,
	"70": fdSiberia, "71": fdCentral, "72": fdUral, "73": fdVolga,
	"74": fdUral, "75": fdFarEast, "76": fdCentral, "77": fdCentral,
	"78": fdNorthWest, "79": fdFarEast, "83": fdNorthWest,
	"86": fdUral, "87": fdFarEast, "89": fdUral,
	"90": fdSouth, // Запорожская обл.
	"91": fdSouth, // Респ. Крым
	"92": fdSouth, // г. Севастополь
	"93": fdSouth, // Донецкая Народная Респ.
	"94": fdSouth, // Луганская Народная Респ.
	"95": fdSouth, // Херсонская обл.
}
