package parser

var AllFieldsIndexes = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
var FieldToIndex = map[string]int{
	"ipaddr":    0,
	"garbage":   1,
	"timestamp": 2,
	"method":    3,
	"url":       4,
	"version":   5,
	"code":      6,
	"size":      7,
	"referrer":  8,
	"useragent": 9,
}
