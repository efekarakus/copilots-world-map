// Package countrydb provides databases to store and retrieve country visit stats.
package countrydb

// Country represents a visited country.
type Country struct {
	Country string // name of the country. ex: Turkey.
	Visit int // total number of visits.
}

// DB is a database that can save a new visit to a country, retrieve all countries visited,
// and fetch number of unique countries visited.
type DB interface {
	Save(country string) (int, error)
	Results() ([]Country, error)
	UniqueTotal() (int, error)
}