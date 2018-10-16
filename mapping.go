package bigger

import (
	"fmt"
    "time"
)

type (
    mappingModule struct {
		tttt	coreBranch
		crypto	coreBranch
    }
)

func (module *mappingModule) Type(name string, config Map, overrides ...bool) {
    
    types := []string{}
    if vv,ok := config[kTYPE].(string); ok && vv != "" {
        types = append(types, vv)
    } else if vvs,ok := config[kTYPES].([]string); ok {
        types = append(types, vvs...)
    } else {
        types = append(types, name)
    }

    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    for _,real := range types {
        if override {
            module.tttt.chunking(real, config)
        } else {
            if module.tttt.chunkdata(real) == nil {
                module.tttt.chunking(real, config)
            }
        }
    }

}
func (module *mappingModule) Crypto(name string, config Map, overrides ...bool) {

    cryptos := []string{}
    if vv,ok := config[kCRYPTO].(string); ok && vv != "" {
        cryptos = append(cryptos, vv)
    } else if vvs,ok := config[kCRYPTOS].([]string); ok {
        cryptos = append(cryptos, vvs...)
    } else {
        cryptos = append(cryptos, name)
    }

    override := true
    if len(overrides) > 0 {
        override = overrides[0]
    }

    for _,real := range cryptos {
        if override {
            module.crypto.chunking(real, config)
        } else {
            if module.crypto.chunkdata(real) == nil {
                module.crypto.chunking(real, config)
            }
        }
    }
}


func (module *mappingModule) typeDefaultValid(value Any, config Map) bool {
    if t,ok := config[kTYPE]; ok {
        return Bigger.Match(fmt.Sprintf("%s", value), fmt.Sprintf("%v", t))
    }
    return false
}
func (module *mappingModule) typeDefaultValue(value Any, config Map) Any {
    return fmt.Sprintf("%s", value)
}








func (module *mappingModule) typeValid(name string) (func(Any, Map) bool) {
	if config,ok := module.tttt.chunkdata(name).(Map); ok {
		switch method := config[kVALID].(type) {
        case func(Any,Map) bool:
            return method
        }
	}
    return module.typeDefaultValid
}
func (module *mappingModule) typeValue(name string) (func(Any,Map) Any) {
	if config,ok := module.tttt.chunkdata(name).(Map); ok {
		switch method := config[kVALUE].(type) {
        case func(Any,Map) Any:
            return method
        }
	}
    return module.typeDefaultValue
}
func (module *mappingModule) typeMethod(name string) (func(Any, Map) bool, func(Any,Map) Any) {
    return module.typeValid(name), module.typeValue(name)
}






func (module *mappingModule) cryptoDefaultEncode(value Any) Any {
    return value
}
func (module *mappingModule) cryptoDefaultDecode(value Any) Any {
    return value
}





func (module *mappingModule) cryptoEncode(name string) (encode func(Any) Any) {
	if config,ok := module.crypto.chunkdata(name).(Map); ok {
		switch method := config[kENCODE].(type) {
        case func(Any) Any:
            return method
        }
	}
    return module.cryptoDefaultEncode
}
func (module *mappingModule) cryptoDecode(name string) (func(Any) Any) {
	if config,ok := module.crypto.chunkdata(name).(Map); ok {
		switch method := config[kDECODE].(type) {
        case func(Any) Any:
            return method
        }
	}
    return module.cryptoDefaultDecode
}
func (module *mappingModule) cryptoMethod(name string)(func(Any) Any, func(Any) Any) {
    return module.cryptoEncode(name), module.cryptoDecode(name)
}










//解析
// func (module *mappingModule) Parse(tree []string, config Map, data Map, value Map, argn bool, pass bool, ctxs ...*Context) *Error {
func (module *mappingModule) Parse(config Map, data Map, value Map, argn bool, pass bool, ctxs ...*Context) *Error {
    var ctx *Context
    if len(ctxs) > 0 {
        ctx = ctxs[0]
    }

    /*
    argn := false
    if len(args) > 0 {
        argn = args[0]
    }
    */


    //遍历配置	begin
    for fieldName,fv := range config {

        fieldConfig := Map{}

        //注意，这里存在2种情况
        //1. Map对象
        //2. map[string]interface{}
        //要分开处理
        //go1.9以后可以 type xx=yy 就只要处理一个了

        switch c := fv.(type) {
        case Map:
            fieldConfig = c
        default:
            //类型不对，跳过
            continue
        }

        //解过密？
        decoded := false
        passEmpty := false
        passError := false

        //Map 如果是JSON文件，或是发过来的消息，就不支持
        //而用下面的，就算是MAP也可以支持，OK
        //麻烦来了，web.args用下面这样处理不了
        //if fieldConfig, ok := fv.(map[string]interface{}); ok {

        fieldMust, fieldEmpty := false, false
        fieldValue, fieldExist := data[fieldName]
        fieldAuto, fieldJson := fieldConfig[kAUTO], fieldConfig[kJSON]
        //_, fieldEmpty = data[fieldName]

        /* 处理是否必填和为空 */
        if v, ok := fieldConfig[kMUST]; ok {
            if v == nil {
                fieldEmpty = true
            }
            if vv,ok := v.(bool); ok {
                fieldMust = vv
            }
        }


        // trees := append(tree, fieldName)

        //fmt.Printf("t=%s, argn=%v, value=%v\n", strings.Join(trees, "."), argn, fieldValue)
        //fmt.Printf("trees=%v, must=%v, empty=%v, exist=%v, auto=%v, value=%v, config=%v\n\n",
        //	strings.Join(trees, "."), fieldMust, fieldEmpty, fieldExist, fieldAuto, fieldValue, fieldConfig)

        strVal := fmt.Sprintf("%v", fieldValue)

        //等一下。 空的map[]无字段。 需要也表示为空吗?
        //if strVal == "" || strVal == "map[]" || strVal == "{}"{

        //因为go1.6之后。把一个值为nil的map  再写入map之后, 判断 if map[key]==nil 就无效了
        if strVal == "" || data[fieldName] == nil || (fieldMust != true && strVal == "map[]")  {
            fieldValue = nil
        }

        //如果不可为空，但是为空了，
        if fieldMust && fieldEmpty == false && (fieldValue == nil || strVal == "") && fieldAuto == nil && fieldJson == nil && argn == false {

            //是否跳过
            if pass {
                passEmpty = true
            } else {
                //是否有自定义的状态
                if empty,ok := fieldConfig[kEMPTY].(string); ok && empty!="" {
                    //自定义的状态下， 应该不用把参数名传过去了，都自定义了
                    return newError(empty)
                } else if empty,ok := fieldConfig[kEMPTY].(*Error); ok{
                    return empty
                } else {
                    if name,ok := fieldConfig[kNAME].(string); ok && name!="" {
						return newError(".mapping.empty", name)
                    } else {
                        // return newError(".mapping.empty", strings.Join(trees, "."))
                        return newError(".mapping.empty", fieldName)
                    }
                }
            }

        } else {

            //如果值为空的时候，看有没有默认值
            //到这里值是可以为空的
            if fieldValue == nil || strVal == "" {

                //如果有默认值
                //可为NULL时，不给默认值
                if fieldAuto != nil && !argn {

                    //暂时不处理 $id, $date 之类的定义好的默认值，不包装了
                    switch autoValue:=fieldAuto.(type) {
                    case func() interface{}:
                        fieldValue = autoValue()
                    case func() time.Time:
                        fieldValue = autoValue()
                        //case func() bson.ObjectId:	//待处理
                        //fieldValue = autoValue()
                    case func() string:
                        fieldValue = autoValue()
                    case func() int:
                        fieldValue = int64(autoValue())
                    case func() int8:
                        fieldValue = int64(autoValue())
                    case func() int16:
                        fieldValue = int64(autoValue())
                    case func() int32:
                        fieldValue = int64(autoValue())
                    case func() int64:
                        fieldValue = autoValue()
                    case func() float32:
                        fieldValue = float64(autoValue())
                    case func() float64:
                        fieldValue = autoValue()
                    case int: {
                        fieldValue = int64(autoValue)
                    }
                    case int8: {
                        fieldValue = int64(autoValue)
                    }
                    case int16: {
                        fieldValue = int64(autoValue)
                    }
                    case int32: {
                        fieldValue = int64(autoValue)
                    }
                    case float32: {
                        fieldValue = float64(autoValue)
                    }

                    case []int: {
                        arr := []int64{}
                        for _,nv := range autoValue {
                            arr = append(arr, int64(nv))
                        }
                        fieldValue = arr
                    }
                    case []int8: {
                        arr := []int64{}
                        for _,nv := range autoValue {
                            arr = append(arr, int64(nv))
                        }
                        fieldValue = arr
                    }
                    case []int16: {
                        arr := []int64{}
                        for _,nv := range autoValue {
                            arr = append(arr, int64(nv))
                        }
                        fieldValue = arr
                    }
                    case []int32: {
                        arr := []int64{}
                        for _,nv := range autoValue {
                            arr = append(arr, int64(nv))
                        }
                        fieldValue = arr
                    }

                    case []float32: {
                        arr := []float64{}
                        for _,nv := range autoValue {
                            arr = append(arr, float64(nv))
                        }
                        fieldValue = arr
                    }

                    default:
                        fieldValue = autoValue
                    }


                    //默认值是不是也要包装一下，这里只包装值，不做验证
                    if fieldType, ok := fieldConfig[kTYPE].(string); ok {
                        _, fieldValueCall := module.typeMethod(fieldType)

                        //如果配置中有自己的值函数
                        if f, ok := fieldConfig[kVALUE].(func(Any, Map) Any); ok {
                            fieldValueCall = f
                        }

                        //包装值
                        if fieldValueCall != nil {
                            fieldValue = fieldValueCall(fieldValue, fieldConfig)
                        }
                    }


                } else {	//没有默认值, 且值为空

                    //有个问题, POST表单的时候.  表单字段如果有，值是存在的，会是空字串
                    //但是POST的时候如果有argn, 实际上是不想存在此字段的

                    //如果字段可以不存在
                    if fieldEmpty || argn {

                        //当empty(argn)=true，并且如果传过值中存在字段的KEY，值就要存在，以更新为null
                        if argn && fieldExist {
                            //不操作，自然往下执行
                        } else {	//值可以不存在
                            continue
                        }

                    } else if argn {


                    } else {
                        //到这里不管
                        //因为字段必须存在，但是值为空
                    }
                }

            } else {	//值不为空，处理值

                //值处理前，是不是需要解密
                //如果解密哦
                //decode
                if ct,ok := fieldConfig[kDECODE].(string); ok && ct !="" {

                    //有一个小bug这里，在route的时候， 如果传的是string，本来是想加密的， 结果这里变成了解密
                    //还得想个办法解决这个问题，所以，在传值的时候要注意，另外string型加密就有点蛋疼了
                    //主要是在route的时候，其它的时候还ok，所以要在encode/decode中做类型判断解还是不解

                    //而且要值是string类型
                    if sv,ok := fieldValue.(string); ok {

                        //得到解密方法
                        decode := module.cryptoDecode(ct)
                        decodeValue := decode(sv)

                        //不相等，才是解过密
                        if decodeValue != sv {
                            //前方解过密了，表示该参数，不再加密
                            //因为加密解密，只有一个2选1的
                            //比如 args 只需要解密 data 只需要加密
                            //route 的时候 args 需要加密，而不用再解，所以是单次的
                            fieldValue = decodeValue
                            decoded = true
                        }
                    }
                }


                //验证放外面来，因为默认值也要验证和包装


                //按类型来做处理

                //验证方法和值方法
                //但是因为默认值的情况下，值有可能是为空的，所以要判断多一点
                if fieldType, ok := fieldConfig[kTYPE].(string); ok && fieldType !="" {
                    fieldValidCall, fieldValueCall := module.typeMethod(fieldType)

                    //如果配置中有自己的验证函数
                    if f,ok := fieldConfig[kVALID].(func(Any,Map)bool); ok {
						fieldValidCall = f
                    }
                    //如果配置中有自己的值函数
                    if f,ok := fieldConfig[kVALUE].(func(Any,Map)Any); ok {
                        fieldValueCall = f
                    }

                    //开始调用验证
                    if fieldValidCall != nil {
                        //如果验证通过
                        if  fieldValidCall(fieldValue, fieldConfig) {
                            //包装值
                            if fieldValueCall != nil {
                                //对时间值做时区处理
                                if ctx != nil && ctx.Zone != time.Local {
                                    if vv,ok := fieldValue.(time.Time); ok {
                                        fieldValue = vv.In(ctx.Zone)
                                    } else if vvs,ok := fieldValue.([]time.Time); ok {
                                        newTimes := []time.Time{}
                                        for _,vv := range vvs {
                                            newTimes = append(newTimes, vv.In(ctx.Zone))
                                        }
                                        fieldValue = newTimes
                                    }
                                }

                                fieldValue = fieldValueCall(fieldValue, fieldConfig)
                            }
                        } else {	//验证不通过

                            //是否可以跳过
                            if pass {
                                passError = true
                            } else {

                                //是否有自定义的状态
                                if c,ok := fieldConfig[kERROR].(string); ok && c != "" {
                                    //自定义的状态下， 应该不用把参数名传过去了，都自定义了
                                    return newError(c)
                                } else if err,ok := fieldConfig[kERROR].(*Error); ok {
                                    return err
                                } else {
                                    if nn,ok := fieldConfig[kNAME].(string); ok && nn != "" {
                                        return newError(".mapping.error", nn)
                                    } else {
                                        return newError(".mapping.error", fieldName)
                                        // return newError(".mapping.error", strings.Join(trees, "."))
                                    }
                                }
                            }
                        }
                    }
                }


            }






        }

        //这后面是总的字段处理
        //如JSON，加密

        //如果是JSON， 或是数组啥的处理
        //注意，当 json 本身可为空，下级有不可为空的，值为空时， 应该跳过子级的检查
        //如果 json 可为空， 就不应该有 默认值， 定义的时候要注意啊啊啊啊
        //理论上，只要JSON可为空～就不处理下一级json
        jsonning := true
        if !fieldMust && fieldValue == nil {
            jsonning = false
        }

        //还有种情况要处理. 当type=json, must=true的时候,有默认值, 但是没有定义json节点.

        if m,ok := fieldConfig[kJSON]; ok && jsonning {
            jsonConfig := Map{}
            //注意，这里存在2种情况
            //1. Map对象 //2. map[string]interface{}
            switch c := m.(type) {
            case Map:
                jsonConfig = c
            }


            //如果是数组
            isArray := false
            //fieldData到这里定义
            fieldData := []Map{}

            switch v := fieldValue.(type) {
            case Map:
                fieldData = append(fieldData, v)
            case []Map:
                isArray = true
                fieldData = v
            default:
                fieldData = []Map{}
            }


            //直接都遍历
            values := []Map{}

            for _,d := range fieldData {
                v := Map{}

                // err := module.Parse(trees, jsonConfig, d, v, argn, pass);
                err := module.Parse(jsonConfig, d, v, argn, pass);
                if err != nil {
                    return err
                } else {
                    //fieldValue = append(fieldValue, v)
                    values = append(values, v)
                }
            }

            if isArray {
                fieldValue = values
            } else {
                if len(values) > 0 {
                    fieldValue = values[0]
                } else {
                    fieldValue = Map{}
                }
            }

        }


        // 跳过且为空时，不写值
        if pass && passEmpty {
        } else {

            // 跳过但错误时，不编码
            if  pass && passError  {

            } else {

                //当pass=true的时候， 这里可能会是空值，那应该跳过
                //最后，值要不要加密什么的
                //如果加密
                //encode
                if ct,ok := fieldConfig[kENCODE].(string); ok && decoded == false && passEmpty == false && passError == false {

                    /*
                    //全都转成字串再加密
                    //为什么要全部转字串才能加密？
                    //不用转了，因为hashid这样的加密就要int64
                    */

                    //得到解密方法
                    encode := module.cryptoEncode(ct)
                    fieldValue = encode(fieldValue)
                }


            }

            //没有JSON要处理，所以给值
            value[fieldName] = fieldValue
        }

    }
    return nil
    //遍历配置	end
}
