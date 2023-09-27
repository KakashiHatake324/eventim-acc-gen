package module_eventim

import (
	"encoding/json"
	"errors"
	module "eventim-acc-gen/module"
	"eventim-acc-gen/requests"
	"fmt"
	"log"
	"net/http"
)

const (
	inHouseApiURL = "https://api.frisapi.dev/akamai"
	apiKey        = "YOUR_KEY"
	apiIndicator  = "success"
)

func (t *TestStructure) genInHouseSensorData() error {
	log.Println("getting sensor data")
	if err := t.generateAkamaiSensorRequest(); err != nil {
		return err
	}
	jsonData, _ := json.Marshal(t.Akamai.Sensor)
	Request := requests.DoRequest{
		Client:         http.DefaultClient,
		CTX:            t.Context,
		AcceptedStatus: []int{200},
		Req:            map[string]string{"Method": "POST", "URL": fmt.Sprintf("%s/sensor", inHouseApiURL), "Data": string(jsonData)},
		Headers:        map[string][]string{"accept": {"application/json"}, "content-type": {"application/json"}},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error getting sensor data", response.Error.Error())
		return errors.New("network error")
	}
	log.Println("getting sensor data =>", response.RespStatus)
	if response.RespStatus == 200 {
		var akamaiResponse = make(map[string]interface{})
		json.Unmarshal([]byte(response.ResponseBody), &akamaiResponse)
		t.Akamai.SensorData = akamaiResponse["data"].(string)
		return nil
	} else {
		return fmt.Errorf("failed with status code %d", response.RespStatus)
	}
}

func (t *TestStructure) genInHousePixelData() error {
	log.Println("getting pixel data")
	if err := t.generateAkamaiPixelRequest(); err != nil {
		return err
	}
	jsonData, _ := json.Marshal(t.Akamai.Pixel)
	Request := requests.DoRequest{
		Client:         http.DefaultClient,
		CTX:            t.Context,
		AcceptedStatus: []int{200},
		Req:            map[string]string{"Method": "POST", "URL": fmt.Sprintf("%s/pixel", inHouseApiURL), "Data": string(jsonData)},
		Headers:        map[string][]string{"accept": {"application/json"}, "content-type": {"application/json"}},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error getting pixel data =>", response.Error.Error())
		return errors.New("network error")
	}
	log.Println("getting pixel data =>", response.RespStatus)
	if response.RespStatus == 200 {
		var akamaiResponse = make(map[string]interface{})
		json.Unmarshal([]byte(response.ResponseBody), &akamaiResponse)
		t.Akamai.PixelData = akamaiResponse["data"].(string)
		return nil
	} else {
		return fmt.Errorf("failed with status code %d", response.RespStatus)
	}
}

func (t *TestStructure) generateAkamaiSensorRequest() error {
	abck, err := t.FindCookie("_abck")
	if err != nil {
		return err
	}
	bmsz, err := t.FindCookie("bm_sz")
	if err != nil {
		return err
	}
	t.Akamai.Sensor = &module.InhouseSensorAkamai{
		ApiKey:     apiKey,
		Ua:         userAgent,
		PageURL:    t.Headers.Referrer,
		ApiVersion: "static",
		Abck:       abck,
		BmSz:       bmsz,
		Hash:       "none",
	}
	return nil
}

func (t *TestStructure) generateAkamaiPixelRequest() error {
	t.Akamai.Pixel = &module.InhousePixelAkamai{
		ApiKey:    apiKey,
		Ua:        userAgent,
		PageURL:   t.Headers.Referrer,
		ScriptVal: t.Akamai.ScriptVal,
		PixelID:   t.Akamai.Baza,
	}
	return nil
}

func (t *TestStructure) postSensorData() error {
	totalGens++
	log.Println("posting akamai sensor data")
	postData := map[string]string{
		"sensor_data": t.Akamai.SensorData,
	}
	postJson, _ := json.Marshal(postData)
	Request := requests.DoRequest{
		Client:         t.Client,
		CTX:            t.Context,
		AcceptedStatus: []int{200, 201},
		Req:            map[string]string{"Method": "POST", "URL": t.Akamai.AkamWebURL, "Data": string(postJson)},
		Headers: map[string][]string{
			"accept":             {"*/*"},
			"content-type":       {"text/plain;charset=UTF-8"},
			"sec-ch-ua":          {t.Headers.SecUa},
			"sec-ch-ua-platform": {t.Headers.Platform},
			"user-agent":         {t.Headers.UserAgent},
			"Sec-Fetch-Site":     {"same-origin"},
			"Sec-Fetch-Mode":     {"cors"},
			"Sec-Fetch-Dest":     {"empty"},
			"referrer":           {t.Headers.Referrer},
			"origin":             {fmt.Sprintf("https://%s", t.Host)},
			"accept-encoding":    {"gzip, deflate, br"},
			"accept-language":    {"en-US,en;q=0.9"},
		},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error posting akamai sensor data =>", response.Error.Error())
		return response.Error
	}
	log.Println("posting akamai sensor data =>", response.RespStatus)
	switch response.RespStatus {
	case 200, 201:
		return nil
	default:
		return errors.New("error posting akamai sensor data")
	}
}

func (t *TestStructure) postPixelData() error {
	log.Println("posting akamai pixel data")
	Request := requests.DoRequest{
		Client:         t.Client,
		CTX:            t.Context,
		AcceptedStatus: []int{200, 201},
		Req: map[string]string{
			"Method": "POST",
			"URL":    fmt.Sprintf("https://%s/akam/13/pixel_%s", t.Host, t.Akamai.PixelV),
			"Data":   t.Akamai.PixelData,
		},
		Headers: map[string][]string{
			"accept":             {"*/*"},
			"content-type":       {"application/x-www-form-urlencoded"},
			"sec-ch-ua":          {t.Headers.SecUa},
			"sec-ch-ua-platform": {t.Headers.Platform},
			"user-agent":         {t.Headers.UserAgent},
			"Sec-Fetch-Site":     {"same-origin"},
			"Sec-Fetch-Mode":     {"cors"},
			"Sec-Fetch-Dest":     {"empty"},
			"referrer":           {t.Headers.Referrer},
			"origin":             {fmt.Sprintf("https://%s", t.Host)},
			"accept-encoding":    {"gzip, deflate, br"},
			"accept-language":    {"en-US,en;q=0.9"},
		},
	}
	response := Request.MakeRequest()
	if response.Error != nil {
		log.Println("error posting akamai pixel data =>", response.Error.Error())
		return response.Error
	}
	log.Println("posting akamai pixel data =>", response.RespStatus)
	switch response.RespStatus {
	case 200, 201:
		return nil
	default:
		return errors.New("error posting akamai pixel data")
	}
}
