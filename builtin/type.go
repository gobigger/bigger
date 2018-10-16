package builtin

import (
	. "github.com/yatlabs/bigger"
	"time"
	"fmt"
	"strings"
	"strconv"
	"encoding/json"
)

func init() {

	Bigger.Type("any", Map{
		"name": "任意类型", "text": "任意类型",
		"valid": func(value Any, config Map) bool {
			return true
		},
		"value": func(value Any, config Map) Any {
			return value
		},
	}, false)


	Bigger.Type("[any]", Map{
		"name": "Anys类型", "text": "Anys类型",
		"valid": func(value Any, config Map) bool {
			return true
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case Any: {
				return []Any{ v }
			}
			case []Any: {
				return v
			}
			default:
			}
			return []Map{}
		},
	}, false)


	Bigger.Type("map", Map{
		"name": "MAP类型", "text": "MAP类型",
		"valid": func(value Any, config Map) bool {
			switch value.(type) {
			case Map: {
				return true
			}
			case []Map: {
				return true
			}
			default:
			}

			return false
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case Map: {
				return v
			}
			case []Map: {
				if len(v) > 0 {
					return v[0]
				}
			}
			default:
			}
			return Map{}
		},
	}, false)

	
	Bigger.Type("[map]", Map{
		"name": "MAPS类型", "text": "MAPS类型",
		"valid": func(value Any, config Map) bool {
			switch value.(type) {
			case Map: {
				return true
			}
			case []Map: {
				return true
			}
			default:
			}

			return false
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case Map: {
				return []Map{ v }
			}
			case []Map: {
				return v
			}
			default:
			}
			return []Map{}
		},
	}, false)



    //---------- bool begin ----------------------------------
	Bigger.Type("bool", Map{
		"name": "布尔型", "text": "布尔型",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case bool: {
				return true
			}
			case string: {
				if v=="true" || v=="false" || v=="0" || v=="1" || v=="yes" || v=="no" {
					return true
				}
			}
			case int,int8,int16,int32,int64,float32,float64:{
				return true
			}
			default:
			}

			return false
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case bool: {
				return v
			}
			case string: {
				if v=="true" || v=="1" || v=="yes" {
					return true
				} else {
					return false
				}
			}
			case int,int8,int16,int32,int64,float32,float64:{
				s := fmt.Sprintf("%v", v)
				if s == "0" {
					return false
				} else {
					return true
				}
			}
			default:

			}

			return false;
		},
	}, false)


	Bigger.Type("[bool]", Map{
		"name": "布尔型数组", "text": "布尔型数组",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case bool: {
				return true
			}
			case []bool: {
				return true
			}
			case string: {

				if (strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}")) ||
					strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {


					//支持以下几种分隔符
					//"," " " ";"
					sp := ","
					if strings.Index(v, ",") >= 0 {
						sp = ","
					} else if strings.Index(v, " ") >= 0 {
						sp = " "
					} else if strings.Index(v, ";") >= 0 {
						sp = ";"
					} else {
						sp = ","
					}

					arr := strings.Split(v[1:len(v)-1], sp)

					for _,sv := range arr {
						sv = strings.TrimSpace(sv)
						if sv=="" { continue }
						if sv!="t" && sv!="T" && sv!="true" && sv!="TRUE" && sv!="1" &&
							sv!="f" && sv!="F" && sv!="FALSE" && sv!="false" && sv!="0" {
							return false
						}
					}

					return true

				} else {

					if v == "true" || v == "false" || v == "0" || v == "1" || v == "yes" || v == "no" || v == "t" || v == "f" {
						return true
					}
				}
			}
			case []string: {
				for _,s := range v {
					if !(s=="true" || s=="false" || s=="0" || s=="1" || s=="yes" || s=="no" || s=="t" || s=="f") {
						return false
					}
				}
				return true
			}
			default:

			}

			return false
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case bool: {
				return []bool { true }
			}
			case []bool: {
				return v
			}
			case string: {

				//支持postgres数组
				if (strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}")) ||
					(strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]")){


					//支持以下几种分隔符
					//"," " " ";"
					sp := ","
					if strings.Index(v, ",") >= 0 {
						sp = ","
					} else if strings.Index(v, " ") >= 0 {
						sp = " "
					} else if strings.Index(v, ";") >= 0 {
						sp = ";"
					} else {
						sp = ","
					}

					arr := []bool{}

					strArr := strings.Split(v[1:len(v)-1], sp)
					for _,sv := range strArr {
						sv = strings.TrimSpace(sv)
						if sv=="" { continue }

						if sv=="t" || sv=="T" || sv=="true" || sv=="TRUE" || sv=="1" {
							arr = append(arr, true)
						} else {
							arr = append(arr, false)
						}
					}

					return arr

				} else {
					if v=="true" || v=="t" || v=="1" || v=="yes" {
						return []bool { true }
					} else {
						return []bool { false }
					}
				}
			}
			case []string: {
				vvvvv := []bool { }
				for _,s := range v {
					if s=="true" || s=="1" || s=="yes" {
						vvvvv = append(vvvvv, true)
					} else {
						vvvvv = append(vvvvv, false)
					}
				}
				return vvvvv
			}
			default:

			}

			return false;
		},
	}, false)
    //----------- bool end ---------------










    //---------- int begin ----------------------------------


	Bigger.Type("int", Map{
		"name": "整型", "text": "整型",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case int,int32,int64,int8: {
				return true
			}
			case float32, float64:
				return true
			case string: {
				v = strings.TrimSpace(v)
				if _, e := strconv.ParseInt(v, 10, 0); e == nil {
					return true
				}
			}
			default:
			}

			return false
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case int: {
				return int64(v)
			}
			case int8: {
				return int64(v)
			}
			case int16: {
				return int64(v)
			}
			case int32: {
				return int64(v)
			}
			case int64: {
				return int64(v)
			}
			case float32:
				return int64(v)
			case float64:
				return int64(v)
			case string: {
				v = strings.TrimSpace(v)
				if i, e := strconv.ParseInt(v, 10, 0); e == nil {
					return i
				}
			}
			default:

			}

			return int64(0)
		},
	}, false)


	Bigger.Type("[int]", Map{
		"name": "整型数组", "text": "整型数组",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case int,int8,int16,int32,int64: {
				return true
			}
			case []int,[]int8,[]int16,[]int32,[]int64: {
				return true
			}
			case float32, float64:
				return true
			case []float32, []float64:
				return true
			case []string: {

				if len(v) > 0 {
					for _,sv := range v {
						sv = strings.TrimSpace(sv)
						if _, e := strconv.ParseInt(sv, 10, 64); e != nil {
							return false
						}
					}
					return true
				}
			}
		case []Any:
			if len(v) > 0 {
				for _,av := range v {
					sv := strings.TrimSpace(fmt.Sprintf("%v", av))
					if _, e := strconv.ParseInt(sv, 10, 64); e != nil {
						return false
					}
				}
				return true
			}
			case string: {

				//此为postgresql数组返回的数组格式
				//{1,2,3,4,5,6,7,8,9}
				if (strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}")) ||
					strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {


					//支持以下几种分隔符
					//"," " " ";"
					sp := ","
					if strings.Index(v, ",") >= 0 {
						sp = ","
					} else if strings.Index(v, " ") >= 0 {
						sp = " "
					} else if strings.Index(v, ";") >= 0 {
						sp = ";"
					} else {
						sp = ","
					}

					arr := strings.Split(v[1:len(v)-1], sp)

					for _,sv := range arr {
						sv = strings.TrimSpace(sv)
						if sv != "" {
							if _, e := strconv.ParseInt(sv, 10, 64); e != nil {
								return false
							}
						}
					}


					return true
					/*
					//不再使用json解析，因为json解析大数字，18位数时，会有精度问题
				} else if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
					jv := []interface{}{}
					e := json.Unmarshal([]byte(v), &jv)
					if e == nil {
						return true
					} else {
						return false
					}*/
				} else {

					v = strings.TrimSpace(v)
					if _, e := strconv.ParseInt(v, 10, 64); e == nil {
						return true
					} else {
						return false
					}
				}




			}

			default:
			}

			return false
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case int: {
				return []int64{ int64(v) }
			}
			case int8: {
				return []int64{ int64(v) }
			}
			case int16: {
				return []int64{ int64(v) }
			}
			case int32: {
				return []int64{ int64(v) }
			}
			case int64: {
				return []int64{ int64(v) }
			}

			case []int: {
				arr := []int64{}
				for _,iv := range v {
					arr = append(arr, int64(iv))
				}
				return arr
			}
			case []int8: {
				arr := []int64{}
				for _,iv := range v {
					arr = append(arr, int64(iv))
				}
				return arr
			}
			case []int16: {
				arr := []int64{}
				for _,iv := range v {
					arr = append(arr, int64(iv))
				}
				return arr
			}
			case []int32: {
				arr := []int64{}
				for _,iv := range v {
					arr = append(arr, int64(iv))
				}
				return arr
			}
			case []int64: {
				arr := []int64{}
				for _,iv := range v {
					arr = append(arr, int64(iv))
				}
				return arr
			}



			case float32:
				return []int64{ int64(v) }
			case float64:
				return []int64{ int64(v) }




			case []float32:
				arr := []int64{}
				for _,iv := range v {
					arr = append(arr, int64(iv))
				}
				return arr
			case []float64:
				arr := []int64{}
				for _,iv := range v {
					arr = append(arr, int64(iv))
				}
				return arr




			case []string: {
				arr := []int64{}
				for _,sv := range v {
					sv = strings.TrimSpace(sv)
					if iv, e := strconv.ParseInt(sv, 10, 64); e == nil {
						arr = append(arr, iv)
					}
				}

				return arr
			}
			case []Any: {
				arr := []int64{}
				for _,av := range v {
					sv := strings.TrimSpace(fmt.Sprintf("%v", av))
					if iv, e := strconv.ParseInt(sv, 10, 64); e == nil {
						arr = append(arr, int64(iv))
					}
				}

				return arr
			}
			case string: {

				if (strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}")) ||
					(strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]")){

					//支持以下几种分隔符
					//"," " " ";"
					sp := ","
					if strings.Index(v, ",") >= 0 {
						sp = ","
					} else if strings.Index(v, " ") >= 0 {
						sp = " "
					} else if strings.Index(v, ";") >= 0 {
						sp = ";"
					} else {
						sp = ","
					}

					arr := []int64{}
					strArr := strings.Split(v[1:len(v)-1], sp)
					for _,sv := range strArr {
						sv = strings.TrimSpace(sv)
						if sv != "" {
							if iv, e := strconv.ParseInt(sv, 10, 64); e == nil {
								arr = append(arr, iv)
							}
						}
					}
					return arr
					/*
					//不再使用json转换，因为json的float在大数字18位长的时候，会有精度问题，
				} else if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
					jv := []interface{}{}
					e := json.Unmarshal([]byte(v), &jv)

					if e == nil {

						arr := []int64{}
						//所以符合的类型,才写入数组
						//json回转,所有的数都是float64
						for _,anyVal := range jv {
							if newVal,ok := anyVal.(float64); ok {
								arr = append(arr, int64(newVal))
							}
						}

						return arr
					}
					*/
				} else {
					
					v = strings.TrimSpace(v)
					if vvv, e := strconv.ParseInt(v, 10, 64); e == nil {
						return []int64{ vvv }
					}
				}


			}
			default:
			}


			return []int64{}
		},
	}, false)


    //---------- int end ----------------------------------







    //---------- string begin ----------------------------------

	Bigger.Type("string", Map{
		"name": "字符串", "text": "字符串",
		"valid": func(value Any, config Map) bool {
			switch v := value.(type) {
			case string:
				if v != "" {
					return true
				}
			case []byte:
				s := fmt.Sprintf("%s", v)
				if s != "" {
					return true
				}
			default:
				if value != nil {
					return true
				}
			}
			return false
		},
		"value": func(value Any, config Map) Any {
			return strings.TrimSpace(fmt.Sprintf("%v", value))
		},
	}, false)


	Bigger.Type("[string]", Map{
		"name": "字符数组", "text": "字符数组",
		"valid": func(value Any, config Map) bool {
			switch value.(type) {
			case []string:
				//要不要判断是否为空数组
				return true
			case []Any:
				//要不要判断是否为空数组
				return true
			case string:
				return true
				/*
				left, right := v[0:1], v[len(v)-1:len(v)]
				if left == "[" && right == "]" {
					return true
				} else if left == "{" && right == "}" {
					return true
				} else {
					return true
				}
				*/
			default:
				return false
			}
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case []string:

				//去空字串
				reals := []string{}
				for _,sv := range v {
					if sv != "" {
						reals = append(reals, sv)
					}
				}

				return reals
			case []Any:

				//去空字串
				reals := []string{}
				for _,av := range v {
					sv := fmt.Sprintf("%v", av)
					if sv != "" {
						reals = append(reals, sv)
					}
				}

				return reals
			case string:
				left, right := v[0:1], v[len(v)-1:len(v)]
				if strings.HasPrefix(v,`["`) && strings.HasSuffix(v,`"]`) {
					// ["a","b","c"]
					s := v[2:len(v)-2]	//去掉{""}
					if s!="" {
						return strings.Split(s, `","`)
					}
					return []string{}
				} else if left == "[" && right == "]" {

					s := v[1:len(v)-1]	//去掉[] {}
					if s!="" {
						return strings.Split(s, `,`)
					}
					return []string{}

				} else if strings.HasPrefix(v,`{"`) && strings.HasSuffix(v,`"}`) {
					//cockroach字串数组返回格式 {"a","b","c"}
					s := v[2:len(v)-2]	//去掉{""}
					if s!="" {
						return strings.Split(s, `","`)
					}
					return []string{}
				} else if left == "{" && right == "}" {
					//postgresl字符串
					s := v[1:len(v)-1]	//去掉[] {}
					if s!="" {
						return strings.Split(s, `,`)
					}
					return []string{}

				} else {
					return []string{ v }
				}


			/*
			s := v[1:len(v)-1]	//去掉[] {}
			if s == "" {
				return []string{}
			} else {
				return strings.Split(s, ",")
			}
			*/
			default:
				return v
			}
		},
	}, false)




	Bigger.Type("[line]", Map{
		"name": "字符分行数组", "text": "字符分行数组",
		"valid": func(value Any, config Map) bool {
			switch value.(type) {
			case []string:
				return true
			case string:
				return true
			default:
				return false
			}
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case []string:

				//去空字串
				reals := []string{}
				for _,sv := range v {
					sv = strings.TrimSpace(sv)
					if sv != "" {
						reals = append(reals, sv)
					}
				}


				return reals
			case string:

				v = strings.Replace(v, "\r", "", -1)
				arr := strings.Split(v, "\n")


				//去空字串
				reals := []string{}
				for _,sv := range arr {
					sv = strings.TrimSpace(sv)
					if sv != "" {
						reals = append(reals, sv)
					}
				}

				return reals
			default:
				return []string{}
			}
		},
	}, false)
    //---------- string end ----------------------------------



    //---------- datetime begin ----------------------------------

	Bigger.Type("date", Map{
		"name": "日期时间", "text": "日期时间",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case time.Time:
				return true
			case *time.Time:
				return true
			case int64:
				return true
			case string:
				return Bigger.Match(v, "date")
			}

			return false;
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case int64:
				return time.Unix(v, 0)
			case string:
				lay := "2006-01-02"
				if len(v) == 8 {
					lay = "20060102"
				} else if len(v) == 10 {
					lay = "2006-01-02"
				} else {
					lay = "2006-01-02"
				}

				dt,err := time.Parse(lay, v)
				if err == nil {
					return dt
				} else {
					return v
				}
			}

			return value
		},
	}, false)

	Bigger.Type("[date]", Map{
		"name": "日期时间数组", "text": "日期时间数组",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case []time.Time:
				return true
			case *[]time.Time:
				return true
			case string:
				return Bigger.Match(v, "date")
			}

			return false;
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case []time.Time:
				return v
			case *[]time.Time:
				return v
			case string:
				lay := "2006-01-02 15:04:05"
				if len(v) == 8 {
					lay = "20060102"
				} else if len(v) == 10 {
					lay = "2006-01-02"
				} else {
					lay = "2006-01-02 15:04:05"
				}

				dt,err := time.Parse(lay, v)
				if err == nil {
					return []time.Time{dt}
				}
			}

			return value
		},
    }, false)
    

	Bigger.Type("datetime", Map{
		"name": "日期时间", "text": "日期时间",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case time.Time:
				return true
			case *time.Time:
				return true
			case string:
				return Bigger.Match(v, "datetime")
			}

			return false;
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case time.Time:
				return v
			case *time.Time:
				return v
			case string:

				lay := "2006-01-02 15:04:05"
				if len(v) == 8 {
					lay = "20060102"
				} else if len(v) == 10 {
					lay = "2006-01-02"
				} else {
					lay = "2006-01-02 15:04:05"
				}

				dt,err := time.Parse(lay, v)
				if err == nil {
					return dt
				} else {
					return v
				}
			}
			return value
		},
	}, false)

	Bigger.Type("[datetime]", Map{
		"name": "日期时间数组", "text": "日期时间数组",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case []time.Time:
				return true
			case *[]time.Time:
				return true
			case string:
				return Bigger.Match(v, "datetime")
			}

			return false;
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case []time.Time:
				return v
			case *[]time.Time:
				return v
			default:
				return v
			}
		},
    }, false)
    


	Bigger.Type("timestamp", Map{
		"name": "时间戳", "text": "时间戳",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case time.Time:
				return true
			case string:
				return Bigger.Match(v, "datetime")
			}

			return false;
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case time.Time:
				return v.Unix()
			case string:
				dt,err := time.Parse("2006-01-02 15:04:05", v)
				if err == nil {
					return dt.Unix()
				} else {
					return v
				}
			}

			return value;
		},
	}, false)

	Bigger.Type("[timestamp]", Map{
		"name": "时间戳数组", "text": "时间戳数组",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case time.Time:
				return true
			case []time.Time:
				return true
			case string:
				return Bigger.Match(v, "datetime")
			}

			return false;
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case time.Time:
				return []int64{ v.Unix() }
			case []time.Time: {
				ts := []int64{}
				for _,dt := range v {
					ts = append(ts, dt.Unix())
				}
				return ts
			}
			case string:
				//应该JSON解析
				dt,err := time.Parse("2006-01-02 15:04:05", v)
				if err == nil {
					return dt.Unix()
				} else {
					return v
				}
			}

			return value;
		},
	}, false)


    //---------- datetime end ----------------------------------



    //---------- enum begin ----------------------------------

	Bigger.Type("enum", Map{
		"name": "枚举", "text": "枚举",
		"valid": func(value Any, config Map) bool {

			sv := fmt.Sprintf("%v", value)

			if e,ok := config["enum"]; ok {
				for k,_ := range e.(Map) {
					if k == sv {
						return true
					}
				}
			}
			return false
		},
		"value": func(value Any, config Map) Any {
			return fmt.Sprintf("%v", value)
		},
	}, false)


	Bigger.Type("[enum]", Map{
		"name": "枚举数组", "text": "枚举数组",
		"valid": func(value Any, config Map) bool {

			vals := []string{}

			switch v := value.(type) {
			case string: {

				//如果是 {},  []  包括的字串，就做split
				//postgres中的， {a,b,c} 格式
				if strings.HasPrefix(v,`{"`) && strings.HasSuffix(v,`"}`) {
					//cockroach字串数组返回格式 {"a","b","c"}
					s := v[2:len(v)-2]	//去掉{""}
                    vals = strings.Split(s, `","`)
				} else if v[0:1] == "{" && v[len(v)-1:len(v)] == "}" {
                    v = v[1:len(v) - 1]
                    vals = strings.Split(v, ",")
                } else if v[0:1] == "[" && v[len(v)-1:len(v)] == "]" {
					//json数组格式
					json.Unmarshal([]byte(v), &vals)
				} else {
					vals = append(vals, v)
				}
			}
			case []string: {
				vals = v
			}
			default:
				vals = append(vals, fmt.Sprintf("%v", v))
			}


			oks := 0

			//遍历数组， 全部在enum里才行
			if e,ok := config["enum"]; ok {
				for k,_ := range e.(Map) {

					for _,v := range vals {
						if k == v {
							oks++
						}
					}
				}
			}


			if oks >= len(vals) {
				return true
			} else {
				return false
			}

		},
		"value": func(value Any, config Map) Any {
			vals := []string{}

			switch v := value.(type) {
			case string: {

				//如果是 {},  []  包括的字串，就做split
				//postgres中的， {a,b,c} 格式
				//postgres中的， {a,b,c} 格式
				if strings.HasPrefix(v,`{"`) && strings.HasSuffix(v,`"}`) {
					//cockroach字串数组返回格式 {"a","b","c"}
					s := v[2:len(v)-2]	//去掉{""}
					vals = strings.Split(s, `","`)
				} else if v[0:1] == "{" && v[len(v)-1:len(v)] == "}" {
					v = v[1:len(v) - 1]
					vals = strings.Split(v, ",")
				} else if v[0:1] == "[" && v[len(v)-1:len(v)] == "]" {
					//json数组格式
					json.Unmarshal([]byte(v), &vals)
				} else {
					vals = append(vals, v)
				}
			}
			case []string: {
				vals = v
			}
			default:
				vals = append(vals, fmt.Sprintf("%v", v))
			}
			return vals
		},
	}, false)

    //---------- enum end ----------------------------------


    //---------- file begin ----------------------------------


	Bigger.Type("file", Map{
		"name": "file", "text": "file",
		"valid": func(value Any, config Map) bool {

			switch value.(type) {
			case Map:
				return true
			}

			return false
		},
		"value": func(value Any, config Map) Any {


			switch vv := value.(type) {
			case Map:
				return vv
			}
			return value
		},
	}, false)


	Bigger.Type("[file]", Map{
		"name": "文件数组", "text": "文件数组",
		"valid": func(value Any, config Map) bool {

			switch value.(type) {
			case Map:
				return true
			case []Map:
				return true
			}

			return false;
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case Map:
				return []Map{ v }
			case []Map:
				return v
			}
			return []Map{}
		},
	}, false)

    //---------- file end ----------------------------------



    //---------- float begin ----------------------------------

	Bigger.Type("float", Map{
		"name": "浮点型", "text": "布尔型",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case int,int8,int16,int32,int64: {
				return true
			}
			case float32,float64: {
				return true
			}
			default:
				sv := fmt.Sprintf("%v", v)
				sv = strings.TrimSpace(sv)
				if _,e := strconv.ParseFloat(sv, 64); e == nil {
					return true
				}
			}

			return false
		},
		"value": func(value Any, config Map) Any {
			switch v := value.(type) {
			case int: {
				return float64(v)
			}
			case int32: {
				return float64(v)
			}
			case int64: {
				return float64(v)
			}
			case int8: {
				return float64(v)
			}
			case float32: {
				return float64(v)
			}
			case float64: {
				return v
			}
			case string: {
				v = strings.TrimSpace(v)
				if v,e := strconv.ParseFloat(v, 64); e == nil {
					return v
				}
			}
			default:
				sv := fmt.Sprintf("%v", v)
				sv = strings.TrimSpace(sv)
				if v,e := strconv.ParseFloat(sv, 64); e == nil {
					return v
				}
			}

			return float64(0.0)
		},
	}, false)



	Bigger.Type("[float]", Map{
		"name": "浮点数组", "text": "浮点数组",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case int,int8,int16,int32,int64: {
				return true
			}
			case []int,[]int8,[]int16,[]int32,[]int64: {
				return true
			}
			case float32, float64:
				return true
			case []float32, []float64:
				return true
			case []string: {
				if len(v) > 0 {
					for _,sv := range v {
						sv = strings.TrimSpace(sv)
						if _, e := strconv.ParseFloat(sv, 64); e != nil {
							return false
						}
					}
					return true
				}
			}
			case []Any: {
				if len(v) > 0 {
					for _,av := range v {
						sv := strings.TrimSpace(fmt.Sprintf("%v", av))
						if _, e := strconv.ParseFloat(sv, 64); e != nil {
							return false
						}
					}
					return true
				}
			}
			case string: {

				//此为postgresql数组返回的数组格式
				//{1,2,3,4,5,6,7,8,9}
				if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
					s := v[1:len(v)-1]
					if s=="" {
						return true
					}
					arr := strings.Split(s, ",")

					for _,sv := range arr {
						if _, e := strconv.ParseFloat(sv, 64); e != nil {
							return false
						}
					}
					return true
				} else if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
					jv := []interface{}{}
					e := json.Unmarshal([]byte(v), &jv)
					if e == nil {
						return true
					} else {
						return false
					}
				} else {

					if _, e := strconv.ParseFloat(v, 64); e == nil {
						return true
					} else {
						return false
					}
				}




			}
			default:
			}

			return false
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case int: {
				return []float64{ float64(v) }
			}
			case int8: {
				return []float64{ float64(v) }
			}
			case int16: {
				return []float64{ float64(v) }
			}
			case int32: {
				return []float64{ float64(v) }
			}
			case int64: {
				return []float64{ float64(v) }
			}


			case []int: {
				arr := []float64{}
				for _,iv := range v {
					arr = append(arr, float64(iv))
				}
				return arr
			}
			case []int8: {
				arr := []float64{}
				for _,iv := range v {
					arr = append(arr, float64(iv))
				}
				return arr
			}
			case []int16: {
				arr := []float64{}
				for _,iv := range v {
					arr = append(arr, float64(iv))
				}
				return arr
			}
			case []int32: {
				arr := []float64{}
				for _,iv := range v {
					arr = append(arr, float64(iv))
				}
				return arr
			}
			case []int64: {
				arr := []float64{}
				for _,iv := range v {
					arr = append(arr, float64(iv))
				}
				return arr
			}



			case float32:
				return []float64{ float64(v) }
			case float64:
				return []float64{ float64(v) }




			case []float32:
				arr := []float64{}
				for _,iv := range v {
					arr = append(arr, float64(iv))
				}
				return arr
			case []float64:
				return v




			case []string: {
				arr := []float64{}
				for _,sv := range v {
					sv = strings.TrimSpace(sv)
					if iv, e := strconv.ParseFloat(sv, 64); e == nil {
						arr = append(arr, float64(iv))
					}
				}
				return arr
			}

			case []Any: {
				arr := []float64{}
				for _,av := range v {
					sv := strings.TrimSpace(fmt.Sprintf("%v", av))
					if iv, e := strconv.ParseFloat(sv, 64); e == nil {
						arr = append(arr, float64(iv))
					}
				}
				return arr
			}
			case string: {

				if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
					arr := []float64{}
					s := v[1:len(v)-1]
					if s!="" {
						strArr := strings.Split(v[1:len(v)-1], ",")
						for _, sv := range strArr {
							sv = strings.TrimSpace(sv)
							if iv, e := strconv.ParseFloat(sv, 64); e == nil {
								arr = append(arr, iv)
							}
						}
					}
					return arr
				} else if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
					jv := []interface{}{}
					e := json.Unmarshal([]byte(v), &jv)
					if e == nil {

						arr := []float64{}
						//所以符合的类型,才写入数组
						//json回转,所有的数都是float64
						for _,anyVal := range jv {
							if newVal,ok := anyVal.(float64); ok {
								arr = append(arr, newVal)
							}
						}
						return arr
					}
				} else {

					v = strings.TrimSpace(v)
					if vvv, e := strconv.ParseFloat(v, 64); e == nil {
						return []float64{ vvv }
					}
				}


			}
			default:
			}


			return []float64{}
		},
	}, false)

    //---------- float end ----------------------------------




    //---------- image begin ----------------------------------

	Bigger.Type("image", Map{
		"name": "image", "text": "image",
		"valid": func(value Any, config Map) bool {

			switch value.(type) {
			case Map:
				return true
			}

			return false
		},
		"value": func(value Any, config Map) Any {


			switch vv := value.(type) {
			case Map:
				return vv
			}
			return Map{}
		},
	}, false)


	Bigger.Type("[image]", Map{
		"name": "image数组", "text": "image数组",
		"valid": func(value Any, config Map) bool {

			switch value.(type) {
			case Map:
				return true
			case []Map:
				return true
			}

			return false
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case Map:
				return []Map{ v }
			case []Map:
				return v
			}
			return []Map{}
		},
	}, false)
    //---------- image end ----------------------------------



    //---------- json begin ----------------------------------

	Bigger.Type("json", Map{
		"name": "JSON", "text": "JSON",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case Map:
				return true
			case string:
				m := Map{}
				err := json.Unmarshal([]byte(v), &m)
				if err == nil {
					return true
				}
			}

			return false
		},
		"value": func(value Any, config Map) Any {

			switch vv := value.(type) {
			case Map:
				return vv
			case string:
				m := Map{}
				err := json.Unmarshal([]byte(vv), &m)
				if err == nil {
					return m
				}
			}
			return value
		},
	}, false)


	Bigger.Type("[json]", Map{
		"name": "JSON数组", "text": "JSON数组",
		"valid": func(value Any, config Map) bool {

			switch v := value.(type) {
			case Map:
				return true
			case []Map:
				return true
			case []interface{}: //而是这个
				return true
			case string:
				m := []Map{}
				err := json.Unmarshal([]byte(v), &m)
				if err == nil {
					return true
				}
			}

			return false
		},
		"value": func(value Any, config Map) Any {

			switch v := value.(type) {
			case Map:
				return []Map{ v }
			case []Map:
				return v
			case []Any: //而是这个
				vvvv := []Map{}
				for _,m := range v {
					if mv,ok := m.(Map); ok {
						mm := Map{}
						for kkk,vvv := range mv {
							mm[kkk] = vvv
						}
						vvvv = append(vvvv, mm)
					}
				}
				return vvvv

			case string:
				m := []Map{}
				err := json.Unmarshal([]byte(v), &m)
				if err == nil {
					return m
				}
			}
			return []Map{}
		},
	}, false)

	

}