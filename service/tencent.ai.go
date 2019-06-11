package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const faceDetectionURL = "https://api.ai.qq.com/fcgi-bin/face/face_detectface"

type coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type aiQQFaceDetectionResponseFaceShape struct {
	FaceProfile  []coordinate `json:"face_profile"`
	LeftEye      []coordinate `json:"left_eye"`
	RightEye     []coordinate `json:"right_eye"`
	LeftEyebrow  []coordinate `json:"left_eyebrow"`
	RightEyebrow []coordinate `json:"right_eyebrow"`
	Mouth        []coordinate `json:"mouth"`
	Nose         []coordinate `json:"nose"`
}

type aiQQFaceDetectionResponseFaceItem struct {
	FaceID     string                             `json:"face_id"`
	X          int                                `json:"x"`
	Y          int                                `json:"y"`
	Width      int                                `json:"width"`
	Height     int                                `json:"height"`
	Gender     int                                `json:"gender"`
	Age        int                                `json:"age"`
	Expression int                                `json:"expression"`
	Beauty     int                                `json:"beauty"`
	Glass      int                                `json:"glass"`
	Pitch      int                                `json:"pitch"`
	Yaw        int                                `json:"yaw"`
	Roll       int                                `json:"roll"`
	FaceShape  aiQQFaceDetectionResponseFaceShape `json:"face_shape"`
}

type AIQQFaceDetectionResponseData struct {
	ImageWidth  int                                 `json:"image_width"`
	ImageHeight int                                 `json:"image_height"`
	FaceList    []aiQQFaceDetectionResponseFaceItem `json:"face_list"`
}

type aiQQFaceDetectionResponse struct {
	Return  int                           `json:"ret"`
	Message string                        `json:"msg"`
	Data    AIQQFaceDetectionResponseData `json:"data"`
}

type TencentFaceDetectionService struct {
	AppID  string
	AppKey string
	Image  []byte
}

func (s TencentFaceDetectionService) setRequestSign(params map[string]string) {
	params["app_id"] = s.AppID
	params["time_stamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	params["nonce_str"] = RandStringRunes(12)

	names := make([]string, 0, len(params))
	for name := range params {
		names = append(names, name)
	}

	sort.Strings(names)

	u := url.Values{}

	for _, name := range names {
		u.Add(name, params[name])
	}

	result := u.Encode() + "&app_key=" + s.AppKey
	sign := getMD5Hash(result)

	params["sign"] = strings.ToUpper(sign)
	return
}

func (s TencentFaceDetectionService) Compress() ([]byte, error) {
	img, err := jpeg.Decode(bytes.NewReader(s.Image))
	if err != nil {
		return nil, err
	}

	output := new(bytes.Buffer)

	if err := jpeg.Encode(output, img, &jpeg.Options{
		Quality: 60,
	}); err != nil {
		return nil, err
	}

	outputByte := output.Bytes()

	return outputByte, nil
}

func (s TencentFaceDetectionService) ToBase64Image(data []byte) string {
	photoBase64 := base64.StdEncoding.EncodeToString([]byte(data))

	return photoBase64
}

func (s TencentFaceDetectionService) Request() (AIQQFaceDetectionResponseData, error) {

	compressedImage, err := s.Compress()

	if err != nil {
		return AIQQFaceDetectionResponseData{}, err
	}

	image := s.ToBase64Image(compressedImage)

	requestData := map[string]string{
		"image": image,
		"mode":  "1",
	}

	s.setRequestSign(requestData)
	requestDataBytes, err := json.Marshal(requestData)
	if err != nil {
		return AIQQFaceDetectionResponseData{}, err
	}

	request, err := http.NewRequest(http.MethodPost, faceDetectionURL, bytes.NewBuffer(requestDataBytes))

	if err != nil {
		return AIQQFaceDetectionResponseData{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(request)

	if err != nil {
		return AIQQFaceDetectionResponseData{}, err
	}
	defer resp.Body.Close()

	var responseData aiQQFaceDetectionResponse
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(respBytes, &responseData); err != nil {
		return AIQQFaceDetectionResponseData{}, err
	}

	if responseData.Return != 0 {
		return AIQQFaceDetectionResponseData{}, errors.New(responseData.Message)
	}

	return responseData.Data, nil
}
