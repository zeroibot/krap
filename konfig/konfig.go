// Package konfig contains app config related functions and types
package konfig

import (
	"strings"

	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/fn/dyn"
	"github.com/zeroibot/fn/fail"
	"github.com/zeroibot/fn/lang"
	"github.com/zeroibot/fn/number"
	"github.com/zeroibot/fn/str"
	"github.com/zeroibot/rdb"
	"github.com/zeroibot/rdb/ze"
)

const (
	keyGlue  string = "."
	listGlue string = "|"
)

// Initialize the config package
func Initialize() error {
	errs := make([]error, 0)

	appFeatures = make(dict.BoolMap)
	scopedFeatures = make(map[string]dict.StringListMap)

	KVSchema = ze.AddSchema(&KV{}, "config_app", errs)
	Features = ze.AddSchema(&Feature{}, "config_features", errs)
	ScopedFeatures = ze.AddSharedSchema(&ScopedFeature{}, errs)

	if len(errs) > 0 {
		return fail.FromErrors("konfig.Initialize", errs)
	}

	return nil
}

// Load Config lookup from database
func Lookup(rq *ze.Request, appKeys []string) (dict.StringMap, error) {
	if KVSchema == nil {
		return nil, ze.ErrMissingSchema
	}
	kv := KVSchema.Ref
	q := rdb.NewLookupQuery[KV](KVSchema.Table, &kv.Key, &kv.Value)
	q.Where(rdb.In(&kv.Key, appKeys))
	lookup, err := q.Lookup(rq.DB)
	if err != nil {
		rq.AddLog("Failed to load app config from db")
		rq.Status = ze.Err500
		return nil, err
	}
	return lookup, nil
}

// Decorates a Config object with the contents of lookup
func Create[T any](cfg *T, lookup dict.StringMap, defaults *Defaults) *T {
	for key := range defaults.UintMap {
		value := uintOrDefault(lookup, defaults.UintMap, key)
		dyn.SetFieldValue(cfg, getKey(key), value)
	}
	for key := range defaults.IntMap {
		value := intOrDefault(lookup, defaults.IntMap, key)
		dyn.SetFieldValue(cfg, getKey(key), value)
	}
	for key := range defaults.StringMap {
		value := stringOrDefault(lookup, defaults.StringMap, key)
		dyn.SetFieldValue(cfg, getKey(key), value)
	}
	for key := range defaults.StringListMap {
		value := stringListOrDefault(lookup, defaults.StringListMap, key)
		dyn.SetFieldValue(cfg, getKey(key), value)
	}
	return cfg
}

// Extract the second part of <Domain>.<Key>
func getKey(fullKey string) string {
	parts := strings.Split(fullKey, keyGlue)
	if len(parts) != 2 {
		return fullKey
	}
	return parts[1]
}

// Tries to convert lookup[key] to uint, fallsback to defaultValue[key]
func uintOrDefault(lookup dict.StringMap, defaultValue map[string]uint, key string) uint {
	value := defaultValue[key]
	if lookupValue, ok := lookup[key]; ok {
		value = uint(number.ParseInt(lookupValue))
	}
	return value
}

// Tries to convert lookup[key] to int, fallsback to defaultValue[key]
func intOrDefault(lookup dict.StringMap, defaultValue map[string]int, key string) int {
	value := defaultValue[key]
	if lookupValue, ok := lookup[key]; ok {
		value = number.ParseInt(lookupValue)
	}
	return value
}

// Tries to get lookup[key], fallsback to defaultValue[key]
func stringOrDefault(lookup dict.StringMap, defaultValue dict.StringMap, key string) string {
	value, ok := lookup[key]
	return lang.Ternary(ok, value, defaultValue[key])
}

// Tries to convert lookup[key] to []string, fallsback to defaultValue[key]
func stringListOrDefault(lookup dict.StringMap, defaultValue dict.StringListMap, key string) []string {
	value, ok := lookup[key]
	return lang.Ternary(ok, str.CleanSplit(value, listGlue), defaultValue[key])
}
