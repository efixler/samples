package models

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Persistent interface {
    Store() error
    Update() error
    Load() (*Persistent, error)
}

type Validateable interface {
    Validate() error
}

var (
	pMapCache = make(map[string]map[string]string) 
	pLock = new(sync.RWMutex)
)

func mapifyFieldNameMap(val interface{}) (map[string]string, string) {
	elem := reflect.TypeOf(val).Elem()
	tn := elem.Name()
	pLock.RLock()
	rval, exists := pMapCache[tn]
	if (exists) {
		defer pLock.RUnlock()
		return rval, tn
	}
	pLock.RUnlock()
	pLock.Lock()
	defer pLock.Unlock()
	rval = make(map[string]string)
	for i := 0; i < elem.NumField(); i++ {
		tf := elem.Field(i)
		if ! isPortable(tf.Type) { continue }
		pname, destname := tf.Name, tf.Name
		tag := tf.Tag
		if jtag,ok := tag.Lookup("json"); ok {
			destname = strings.SplitN(jtag, ",", 2)[0]
		}
		rval[pname] = destname
	}
	pMapCache[tn] = rval
	return rval, tn
}

func isPortable(t reflect.Type) bool {
	kind := t.Kind()
	//see https://golang.org/pkg/reflect/#Kind iota
	if (kind >= reflect.Bool) && (kind <= reflect.Uint64) {
		return true
	} else if (kind >= reflect.Float32) && (kind <= reflect.Complex128) {
		return true
	} else if (kind == reflect.String) {
		return true
	} else if (kind == reflect.Struct) {
		switch t.Name() {
			case "Time": return true
		}
	}
	return false
}

func toAttributeMap(it interface{}) map[string]string {
	fieldNames, _ := mapifyFieldNameMap(it)
	rval := make(map[string]string)
	val := reflect.ValueOf(it).Elem()
	for skey, tkey := range fieldNames {
		v := val.FieldByName(skey)
		if (! v.IsValid()) {
			continue
		} else if (v.Interface() == reflect.Zero(v.Type()).Interface()) {
			continue
		} else { 
			attr := fmt.Sprintf("%v", v.Interface())
			rval[tkey] = attr
		}
	}
	return rval
}
