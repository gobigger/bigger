package bigger


func (bigger *bigger) builtin() {

	Bigger.Crypto("string", Map{
		"name": "文本加密", "text": "文本加密，自定义字符表的base64编码，字典："+encodeTextAlphabet,
		"encode": func(value Any) Any {
			text := Bigger.ToString(value)
			return Bigger.Decrypt(text)
		},
		"decode": func(value Any) Any {
            text := Bigger.ToString(value)
			return Bigger.Decrypt(text)
		},
	})
    Bigger.Crypto("strings", Map{
		"name": "文本数组加密", "text": "文本数组加密，自定义字符表的base64编码，字典："+encodeTextAlphabet,
		"encode": func(value Any) Any {
			if vv,ok := value.([]string); ok {
				return Bigger.Encrypts(vv)
			}
            return value
		},
		"decode": func(value Any) Any {
			text := Bigger.ToString(value)
			return Bigger.Decrypts(text)
		},
	})



	Bigger.Crypto("number", Map{
		"name": "数字加密", "text": "数字加密",
		"encode": func(value Any) Any {
			if vv,ok := value.(int64); ok {
				return Bigger.Enhash(vv)
			}
			return value
		},
		"decode": func(value Any) Any {
			if vv,ok := value.(string); ok {
				return Bigger.Dehash(vv)
			}
			return value
		},
	})

	Bigger.Crypto("numbers", Map{
		"name": "数字数组加密", "text": "数字数组加密",
		"encode": func(value Any) Any {
			if vv,ok := value.([]int64); ok {
				return Bigger.Enhashs(vv)
            }
            return value
        },
		"decode": func(value Any) Any {
			if vv,ok := value.(string); ok {
				return Bigger.Dehashs(vv)
			}
			return value
		},
	})



}


