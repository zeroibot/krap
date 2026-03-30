package cache

var useCache bool = true

// Toggle usage of cache package
func Use(flag bool) {
	useCache = flag
}
