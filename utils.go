
package utils


var FixedMimeList = map[string]string{
	"js" : "application/x-javascript",
}

var VariantMimeList = map[string]string{
	"application/javascript" : FixedMimeList["js"],
}


func FixedMime(mimeType string) string {
	if v, ok := VariantMimeList[mimeType]; ok{
		return v
	}
	return mimeType
}
