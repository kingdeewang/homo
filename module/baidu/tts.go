//
// Copyright (c) 2019-present Codist <countstarlight@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// Written by Codist <countstarlight@gmail.com>, March 2019
//

package baidu

import (
	"fmt"
	"github.com/countstarlight/homo/cmd/webview/config"
	"github.com/countstarlight/homo/module/audio"
	"github.com/countstarlight/homo/module/com"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
)

//TextToSpeech 对接baidu tts rest api
//https://ai.baidu.com/docs#/TTS-API/top
func (vc *VoiceClient) TextToSpeech(txt string) ([]byte, error) {

	if len(txt) >= 1024 {
		return nil, fmt.Errorf("Input text too long: %d > 1024", len(txt))
	}
	if err := vc.Auth(); err != nil {
		return nil, err
	}

	var cuid string
	netitfs, err := net.Interfaces()
	if err != nil {
		cuid = "anonymous"
	} else {
		for _, itf := range netitfs {
			if cuid = itf.HardwareAddr.String(); len(cuid) > 0 {
				break
			}
		}
	}

	resp, err := http.PostForm(config.BaiduTTSAPI, url.Values{
		"tex":  {txt},
		"tok":  {vc.AccessToken},
		"cuid": {cuid},
		"ctp":  {"1"},
		"lan":  {"zh"},
		"spd":  {"5"},
		"pit":  {"5"},
		"vol":  {"5"},
		"per":  {"0"}, //0: default female, 1: default male, 4: girl
		"aue":  {"6"}, //3: mp3 format 6: wav
	})
	if err != nil {
		return nil, err
	}
	defer com.IOClose("Post baidu tts api", resp.Body)

	//通过Content-Type的头部来确定是否服务端合成成功。
	//http://ai.baidu.com/docs#/TTS-API/top
	respHeader := resp.Header
	contentType, ok := respHeader["Content-Type"]
	if !ok {
		return nil, fmt.Errorf("No Content-Type Set.")
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	/*
		if contentType[0] == "audio/mp3" {
			return respBody, nil
		} else {
			return nil, fmt.Errorf("调用服务失败：%s", string(respBody))
		}
	*/
	if contentType[0] == "audio/wav" {
		return respBody, nil
	} else {
		return nil, fmt.Errorf("调用服务失败：%s", string(respBody))
	}

}

const (
	B = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
)

type VoiceClient struct {
	*Client
}

func NewVoiceClient(apiKey, apiSecret string) *VoiceClient {
	return &VoiceClient{
		Client: NewClient(apiKey, apiSecret),
	}
}

// Voice Composition
func TextToSpeech(text string) error {
	client := NewVoiceClient(config.BaiduVoiceAPIKey, config.BaiduVoiceAPISecret)
	voiceData, err := client.TextToSpeech(text)
	if err != nil {
		return err
	}

	//Remove previous file
	if com.IsFile(config.TTSOutFile) {
		err = os.Remove(config.TTSOutFile)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(config.TTSOutFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer com.IOClose("Save baidu tts to file", f)
	if _, err := f.Write(voiceData); err != nil {
		return err
	}
	//PortAudio
	//err = audio.PortAudioPlayMp3("tmp/tts/tmp.mp3")
	return audio.BeepPlayWav(config.TTSOutFile)
}
