package valueutils

import (
	"reflect"
)

func MergeMaps(dst, src map[string]interface{}) map[string]interface{} {
	return merge(dst, src, 0)
}

func merge(dst, src map[string]interface{}, depth int) map[string]interface{} {
	for key, srcVal := range src {
		if srcVal == nil {
			dst[key] = nil
			continue
		}
		if dstVal, ok := dst[key]; ok {
			srcMap, srcMapOk := mapify(srcVal)
			dstMap, dstMapOk := mapify(dstVal)
			if srcMapOk && dstMapOk {
				srcVal = merge(dstMap, srcMap, depth+1)
			} else if dstVal == nil {
				// pass
			} else if reflect.TypeOf(srcVal).Kind() == reflect.Slice && reflect.TypeOf(dstVal).Kind() == reflect.Slice {
				dstSlice := dstVal.([]interface{})
				var r []interface{}
				for i, item := range srcVal.([]interface{}) {
					if item != nil {
						r = append(r, item)
					} else {
						if len(dstSlice) >= i {
							r = append(r, dstSlice[i])
						} else {
							r = append(r, nil)
						}
					}
				}
				srcVal = r
			}
		}
		dst[key] = srcVal
	}
	return dst
}

func mapify(i interface{}) (map[string]interface{}, bool) {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			m[k.String()] = value.MapIndex(k).Interface()
		}
		return m, true
	}
	return map[string]interface{}{}, false
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}

		// fmt.Println(reflect.TypeOf(v), reflect.TypeOf(v).Kind(), "value for v:", v)
		//
		if v != nil && reflect.TypeOf(v).Kind() == reflect.Slice {
			var r []interface{}
			for i, item := range v.([]interface{}) {
				if item != nil {
					r = append(r, item)
				} else {
					r = append(r, out[k].([]interface{})[i])
				}
			}
			out[k] = r
		} else {
			out[k] = v
		}

	}
	return out
}
