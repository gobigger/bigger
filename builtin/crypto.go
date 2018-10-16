package builtin

import (
	"encoding/base64"
	. "github.com/yatlabs/bigger"
)

func init() {

	Bigger.Crypto("base64", Map{
		"name": "BASE64加解密", "text": "BASE64加解密",
		"encode": func(value Any) Any {
			text := Bigger.ToString(value)
			return base64.StdEncoding.EncodeToString([]byte(text))
		},
		"decode": func(value Any) Any {
			text := Bigger.ToString(value)
			bytes,err := base64.StdEncoding.DecodeString(text)
			if err == nil {
				return string(bytes)
			}
			return value
		},
	}, false)

}


