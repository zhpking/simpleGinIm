package helper

import "github.com/fatih/structs"

func Struct2MapByKey(key string, s []interface{}) map[string]interface{} {
	ret := map[string]interface{}{}
	for _, val := range s {
		s := structs.New(val)
		value := s.Field(key).Value().(string)
		ret[value] = val
	}

	return ret
}

func StructList2MaoListByKey(key string, s []interface{}) map[string][]interface{} {
	ret := map[string][]interface{}{}
	for _, val := range s {
		s := structs.New(val)
		value := s.Field(key).Value().(string)
		if _, ok := ret[value]; !ok {
			ret[value] = make([]interface{}, 0)
		}

		ret[value] = append(ret[value], val)
	}

	return ret
}

/*
room
1 => [user=>1,room=>1]
2 => [user=>2,room=>2]
3 => [user=>3,room=>3]

message
1 => [message=>1,room=>1]
2 => [message=>2,room=>2]
3 => [message=>3,room=>3]
*/

func combineStructList(mainData []interface{}, combineKey string, m ...map[string]interface{}) []map[string]interface{} {
	retList := make([]map[string]interface{}, 0)
	for i := 0; i < len(mainData); i ++ {
		main := structs.New(mainData[i])
		mainKey := main.Name()
		combileValue := main.Field(combineKey).Value().(string)

		// 单个组合数据
		ret := map[string]interface{}{
			mainKey:mainData[i],
		}
		for _, v := range m {
			if val, ok := v[combileValue]; ok {
				ret[structs.New(val).Name()] = val
			}
		}

		retList = append(retList, ret)
	}

	return retList
}
