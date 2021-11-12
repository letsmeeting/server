package utility

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// ParseTokenInt
// token으로 문자열을 분리하여 integer slice로 반환하며 공백은 모두 무시한다
//  example)
//	s = "1234, 567, 1  "
//	token = ","
//	return = [1234] [567] [1]
func ParseTokenInt(s string, token string) []int {
	list := make([]int, 0)
	s = strings.Trim(s, " ")
	split := strings.Split(s, token)
	for i, _ := range split {
		split[i] = strings.TrimSpace(split[i])
		portNum, err := strconv.Atoi(split[i])
		if err == nil {
			list = append(list, portNum)
		}
	}
	return list
}

// GetPrettyJsonStr
// byte array 를 보기 좋게 정렬된 json string 으로 반환
func GetPrettyJsonStr(message []byte) (string, error) {
	var logBuff bytes.Buffer
	err := json.Indent(&logBuff, message, "", "   ")
	if err != nil {
		return "", err
	}
	logBuff.Write([]byte("\n"))
	return logBuff.String(), nil
}

func DeleteEmpty (s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}