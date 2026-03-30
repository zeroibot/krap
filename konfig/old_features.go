package konfig

import (
	"errors"
	"slices"
	"strings"

	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/rdb"
	"github.com/zeroibot/rdb/ze"
)

var (
	errUnavailableFeature = errors.New("public: Unavailable feature")
	errUnknownFeature     = errors.New("public: Unknown feature")
)

var (
	Features       *ze.Schema[Feature]
	ScopedFeatures *ze.Schema[ScopedFeature]
)

type Feature struct {
	ze.ActiveItem
	Name string `fx:"upper" col:"Feature"`
}

type ScopedFeature struct {
	Feature
	ScopeCode string `fx:"upper"`
}

var (
	appFeatures    dict.BoolMap
	scopedFeatures map[string]dict.StringListMap
)

// Load app features
func LoadFeatures(rq *ze.Request) error {
	if Features == nil {
		return ze.ErrMissingSchema
	}
	f := Features.Ref
	q := rdb.NewLookupQuery[Feature](Features.Table, &f.Name, &f.IsActive)
	lookup, err := q.Lookup(rq.DB)
	if err != nil {
		rq.AddLog("Failed to load features from db")
		rq.Status = ze.Err500
		return err
	}
	appFeatures = lookup
	return nil
}

// Load active scoped features at table
func LoadScopedFeatures(rq *ze.Request, table string) error {
	if ScopedFeatures == nil {
		return ze.ErrMissingSchema
	}

	f := ScopedFeatures.Ref
	condition := rdb.Equal(&f.IsActive, true)
	features, err := ScopedFeatures.GetRowsAt(rq, condition, table)
	if err != nil {
		rq.AddFmtLog("Failed to load scoped features from '%s'", table)
		return err
	}

	if scopedFeatures == nil {
		scopedFeatures = make(map[string]dict.StringListMap)
	}
	scopedFeatures[table] = make(dict.StringListMap)

	for _, f := range features {
		scope, feature := f.ScopeCode, f.Name
		scopedFeatures[table][scope] = append(scopedFeatures[table][scope], feature)
	}

	return nil
}

// Get map[Feature]IsActive
func GetAllFeatures() dict.BoolMap {
	return appFeatures
}

// Get list of active app features
func GetActiveFeatures() []string {
	if appFeatures == nil {
		return []string{}
	}
	activeFeatures := dict.Filter(appFeatures, func(feature string, isActive bool) bool {
		return isActive
	})
	return dict.Keys(activeFeatures)
}

// Get all active {scope => []features} at table
func GetAllScopedFeatures(table string) dict.StringListMap {
	if dict.NoKey(scopedFeatures, table) {
		return dict.StringListMap{}
	}
	return scopedFeatures[table]
}

// Get all active {feature => []scopes} at table
func GetAllFeatureScopes(table string) dict.StringListMap {
	featureScopes := make(dict.StringListMap)
	for scope, features := range scopedFeatures[table] {
		for _, feature := range features {
			featureScopes[feature] = append(featureScopes[feature], scope)
		}
	}
	for feature, scopes := range featureScopes {
		slices.Sort(scopes)
		featureScopes[feature] = scopes
	}
	return featureScopes
}

// Get all active scoped features at table for given scopeCodes
func GetScopedFeatures(table string, scopeCodes ...string) dict.StringListMap {
	scopeFeatures := make(dict.StringListMap)
	if dict.NoKey(scopedFeatures, table) {
		return scopeFeatures
	}
	for _, scope := range scopeCodes {
		scope = strings.ToUpper(scope)
		features := scopedFeatures[table][scope]
		if len(features) == 0 {
			features = []string{}
		}
		slices.Sort(features)
		scopeFeatures[scope] = features
	}
	return scopeFeatures
}

// Check if feature is available
func CheckFeature(rq *ze.Request, feature string) error {
	feature = strings.ToUpper(feature)
	isActive, ok := appFeatures[feature]
	if !ok {
		rq.Status = ze.Err404
		return errUnknownFeature
	}
	if !isActive {
		rq.Status = ze.Err403
		return errUnavailableFeature
	}
	return nil
}

// Check if scoped feature at table is available
func CheckScopedFeature(rq *ze.Request, table, scopeCode, feature string) error {
	if dict.NoKey(scopedFeatures, table) {
		rq.Status = ze.Err403
		return errUnavailableFeature
	}
	scopeCode = strings.ToUpper(scopeCode)
	feature = strings.ToUpper(feature)
	enabled := scopedFeatures[table][scopeCode]
	if !slices.Contains(enabled, feature) {
		rq.Status = ze.Err403
		return errUnavailableFeature
	}
	return nil
}
