package module_eventim

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
)

// Decode a base64 encoded string
func decodeBase64(encoded string) string {
	decoded, _ := base64.StdEncoding.DecodeString(encoded)
	return string(decoded)
}

// Get the value between a string
func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func ParseWrgen(body string) (string, string) {

	var siteKey, action string
	exp := regexp.MustCompile(`(?m)(?m)var K=M(.*)":"6`)
	for _, match := range exp.FindAllString(string(body), -1) {
		rep := strings.NewReplacer(`var K=M()?"`, "", `":"6`, "")
		siteKey = rep.Replace(match)
	}

	var exp2 = regexp.MustCompile(`(?m)(?m)(?m){action:"(.*)"}\);`)
	for _, match2 := range exp2.FindAllString(string(body), -1) {
		rep := strings.NewReplacer(`{action:"`, "", `"});`, "")
		action = rep.Replace(match2)
	}

	return siteKey, action
}

func ParseHomePage(body string) (string, string, string, *ParsingErrors) {

	var bazaValue, pixelTVal, pixelScriptVal string

	bazaCheck, err := regexp.MatchString("bazadebezolkohpepadr", body)

	if err != nil {
		return "", "", "", &ParsingErrors{Err: "Error in Parsing Home page"}
	}

	if bazaCheck {
		bazaValue = between(body, `bazadebezolkohpepadr="`, `"</script>`)
	} else {
		bazaValue = "false"
	}

	pixelCheck, err := regexp.MatchString("akam/13/", body)

	if err != nil {
		return "", "", "", &ParsingErrors{Err: "Error in Parsing Home page"}
	}

	if pixelCheck {
		pixelScriptVal = between(body, `akam/13/`, `"  defer></script></head>`)
		pixelTVal = decodeBase64(between(body, `akam/11/pixel_`+pixelScriptVal+"?a=", `" style="visibility`))
	} else {
		pixelScriptVal = "false"
		pixelTVal = "false"
	}

	return bazaValue, pixelTVal, pixelScriptVal, nil
}

func ParsePixelScript(body string) (string, *ParsingErrors) {

	var gIndex, gVal string
	exp := regexp.MustCompile(`(?m)g=_(.*),m`)

	for _, match := range exp.FindAllString(body, -1) {
		exp := regexp.MustCompile("[0-9]+")
		gIndex = exp.FindAllString(match, -1)[0]
	}

	exp2 := regexp.MustCompile(`(?m)var _=[ []"(.*)];`)
	for _, match2 := range exp2.FindAllString(body, -1) {
		rep := strings.NewReplacer("var _ = [", "", "];", "", `"`, "", "\u0020", "")
		res := rep.Replace(match2)
		arr := strings.Split(res, ",")

		intVar, err := strconv.Atoi(gIndex)
		if err != nil {
			return "", &ParsingErrors{Err: "Pixel Script Parsing Error"}
		}

		rep2 := strings.NewReplacer("\\", "", "x", "", "", "")
		gVal = rep2.Replace(arr[intVar])
		decodedString, err := hex.DecodeString(gVal)
		if err != nil {
			return "", &ParsingErrors{Err: "Pixel Script Parsing Error"}
		}
		gVal = string(decodedString)

	}

	return gVal, nil
}
